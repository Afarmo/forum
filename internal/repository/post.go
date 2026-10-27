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
	// Posts require an author, while profile pictures and categories are optional.
	// COALESCE maps those optional values to empty strings for the feed model.
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.picture_content, p.created_at, p.deleted_at,
			u.username, COALESCE(u.profile_picture, ''),
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT c.name)
				FROM post_categories pc
				JOIN categories c ON c.id = pc.category_id
				WHERE pc.post_id = p.id
			), '')
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.deleted_at IS NULL`
	args := []any{}

	if categoryID > 0 {
		query += `
			AND EXISTS (
				SELECT 1 FROM post_categories pc_filter
				WHERE pc_filter.post_id = p.id AND pc_filter.category_id = ?
			)`
		args = append(args, categoryID)
	}
	query += ` ORDER BY p.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post

	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.PostID, &post.UserID, &post.Title, &post.Content,
			&post.PictureContent, &post.CreatedAt, &post.DeletedAt, &post.AuthorName,
			&post.AuthorPicture, &post.CategoryNames,
		)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (r *PostRepository) GetPopularPosts(ctx context.Context, since time.Time, limit int) ([]models.PopularPost, error) {
	query := `
		SELECT p.id, p.title, COALESCE(u.username, ''), COUNT(*) AS likes
		FROM posts p
		JOIN users u ON u.id = p.user_id
		JOIN post_reactions pr ON pr.post_id = p.id AND pr.reaction = 1
		WHERE p.created_at >= ?
		GROUP BY p.id
		ORDER BY likes DESC, p.created_at DESC
		LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, since.UTC().Format("2006-01-02 15:04:05"), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.PopularPost
	for rows.Next() {
		var p models.PopularPost
		if err := rows.Scan(&p.PostID, &p.Title, &p.AuthorName, &p.Likes); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *PostRepository) GetPostByUser(ctx context.Context, userID int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.content,
			p.picture_content, p.created_at, p.deleted_at,
			u.username, COALESCE(u.profile_picture, ''),
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT c.name)
				FROM post_categories pc
				JOIN categories c ON c.id = pc.category_id
				WHERE pc.post_id = p.id
			), '')
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.user_id = ?
		AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.PostID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.PictureContent,
			&post.CreatedAt,
			&post.DeletedAt,
			&post.AuthorName,
			&post.AuthorPicture,
			&post.CategoryNames,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (r *PostRepository) UpdatePost(ctx context.Context, postID, userID int, update *models.UpdatePost) error {
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
	args = append(args, postID, userID)
	query := fmt.Sprintf("UPDATE posts SET %s WHERE id = ? AND user_id = ? AND deleted_at IS NULL", strings.Join(extraQuery, ", "))
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

func (r *PostRepository) DeletePost(ctx context.Context, postID *int, userID int) error {
	query := `
		UPDATE posts
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, *postID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *PostRepository) GetPostById(ctx context.Context, postID int) (*models.Post, error) {

	query := `SELECT p.id, p.user_id, p.title, p.content, p.picture_content, p.created_at, p.deleted_at,
			COALESCE(u.username, ''), COALESCE(u.profile_picture, '')
		FROM posts p
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.id = ?`

	var post models.Post
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&post.PostID, &post.UserID, &post.Title, &post.Content, &post.PictureContent, &post.CreatedAt, &post.DeletedAt,
		&post.AuthorName, &post.AuthorPicture)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) SearchPosts(ctx context.Context, search string) ([]models.Post, error) {
	if strings.TrimSpace(search) == "" {
		return r.GetAllPosts(ctx, 0)
	}
	//  FTS5 column weights: prioritize title, then username, then content.
	//  Lower BM25 scores are more relevant, so the best matches come first.
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.picture_content, p.created_at, p.updated_at, p.deleted_at,
			u.username, COALESCE(u.profile_picture, ''),
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT c.name)
				FROM post_categories pc
				JOIN categories c ON c.id = pc.category_id
				WHERE pc.post_id = p.id
			), '')
		FROM POSTS_SEARCH
		JOIN POSTS p ON p.id = POSTS_SEARCH.rowid
		JOIN users u ON u.id = p.user_id
		WHERE POSTS_SEARCH MATCH ?
			AND p.deleted_at IS NULL
		
		ORDER BY bm25(POSTS_SEARCH, 10.0, 1.0, 5.0), p.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.PostID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.PictureContent,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.DeletedAt,
			&post.AuthorName,
			&post.AuthorPicture,
			&post.CategoryNames,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (r *PostRepository) GetLikedPostsByUser(ctx context.Context, userID int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.content,
			p.picture_content, p.created_at, p.deleted_at,
			u.username, COALESCE(u.profile_picture, ''),
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT c.name)
				FROM post_categories pc
				JOIN categories c ON c.id = pc.category_id
				WHERE pc.post_id = p.id
			), '')
		FROM posts p
		JOIN post_reactions pr ON pr.post_id = p.id
		JOIN users u ON u.id = p.user_id
		WHERE pr.user_id = ?
		AND pr.reaction = 1
		AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.PostID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.PictureContent,
			&post.CreatedAt,
			&post.DeletedAt,
			&post.AuthorName,
			&post.AuthorPicture,
			&post.CategoryNames,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}
