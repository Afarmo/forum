package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Afarmo/forum/internal/apperrors"
	"github.com/Afarmo/forum/internal/models"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{
		db: db,
	}
}

func (r *PostRepository) CreatePost(ctx context.Context, post *models.Post, categoryID int) error {
	tx, txErr := r.db.BeginTx(ctx, nil)
	if txErr != nil {
		return apperrors.ErrTransactionStart
	}
	defer tx.Rollback() // if there is error revert changes to the db back to before the changes
	query := `INSERT INTO posts(user_id, title, content, picture_content) VALUES(?,?,?,?)`
	now := time.Now()
	result, err := tx.ExecContext(ctx, query, post.UserID, post.Title, post.Content, post.PictureContent)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return apperrors.ErrDuplicateKey
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	post.PostID = int(id)
	post.CreatedAt = now
	categoryQuery := `INSERT INTO post_categories(post_id, category_id) VALUES(?,?)`
	_, err = tx.ExecContext(ctx, categoryQuery, post.PostID, categoryID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostRepository) GetAllPosts(ctx context.Context, categoryID int) ([]models.Post, error) {
	query := `SELECT id, user_id, title, content, picture_content, created_at FROM posts ORDER BY created_at DESC`
	args := []any{}

	if categoryID > 0 {
		query = `
			SELECT p.id, p.user_id, p.title, p.content, p.picture_content, p.created_at
			FROM posts p
			JOIN post_categories pc ON pc.post_id = p.id
			WHERE pc.category_id = ?
			ORDER BY p.created_at DESC`
		args = append(args, categoryID)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.PostID, &post.UserID, &post.Title, &post.Content, &post.PictureContent, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (r *PostRepository) GetPostByUser(ctx context.Context, userId int) ([]models.Post, error) {
	query := `SELECT id, user_id, title, content, picture_content, created_at FROM posts WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post

	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.PostID, &post.UserID, &post.Title, &post.Content, &post.PictureContent, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (r *PostRepository) UpdatePost(ctx context.Context, userID int, update *models.UpdatePost) error {
	tx, txErr := r.db.BeginTx(ctx, nil)
	if txErr != nil {
		return apperrors.ErrTransactionStart
	}
	defer tx.Rollback() // if there is error revert changes to the db back to before the changes
	now := time.Now()
	extraQuery := []string{"updated_at = ?"}
	args := []any{now}
	if update.Title != nil {
		extraQuery = append(extraQuery, `title = ?`)
		args = append(args, *update.Title)
	}
	if update.Content != nil {
		extraQuery = append(extraQuery, `content = ?`)
		args = append(args, *update.Content)
	}
	if update.PictureContent != nil {
		extraQuery = append(extraQuery, `picture_content = ?`)
		args = append(args, *update.PictureContent)
	}
	args = append(args, userID)
	query := fmt.Sprintf("UPDATE posts SET %s WHERE id = ?", strings.Join(extraQuery, ", "))
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraints failed") {
			return apperrors.ErrDuplicateKey
		}
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return tx.Commit()
}

func (r *PostRepository) DeletePost(ctx context.Context, postID *int) error {
	tx, txErr := r.db.BeginTx(ctx, nil)
	if txErr != nil {
		return apperrors.ErrTransactionStart
	}

	defer tx.Rollback()

	queries := []string{`DELETE FROM comment_reaction WHERE comment_id IN(SELECT id FROM comments WHERE post_id = ?)`, `DELETE FROM comments WHERE post_id = ?`, `DELETE FROM post_reactions WHERE post_id = ?`, `DELETE FROM post_categories WHERE post_id = ?`, `DELETE FROM posts WHERE id = ?`}
	for _, query := range queries {
		_, err := tx.ExecContext(ctx, query, postID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostRepository) SearchPosts(ctx context.Context, search string) ([]models.Post, error) {

	query := `SELECT id, user_id, title, content, picture_content, created_at, updated_at FROM POSTS where title LIKE = ? OR content LIKE = ? ORDER BY created_at DESC`
	search = "%" + search + "%"
	rows, err := r.db.QueryContext(ctx, query, search, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next(){
		var post models.Post
		err := rows.Scan(
			&post.PostID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.PictureContent,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil{
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}
