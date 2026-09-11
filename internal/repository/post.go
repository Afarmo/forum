package repository

import (
	"context"
	"database/sql"
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
	query := `INSERT INTO posts(user_id, Content, picture_content) VALUES(?,?,?)`
	now := time.Now()
	result, err := tx.ExecContext(ctx, query, post.UserId, post.Content, post.PictureContent)
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
	post.PostId = int(id)
	post.CreatedAt = now
	categoryQuery := `INSERT INTO post_categories(post_id, category_id) VALUES(?,?)`
	_, err = tx.ExecContext(ctx, categoryQuery, post.PostId, categoryID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostRepository) GetAllPosts( ctx context.Context)([]models.Post, error){
	query := `SELECT id, user_id, content, picture_content, created_at FROM posts ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil{
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next(){
		var post models.Post
		err := rows.Scan(&post.PostId, &post.UserId, &post.Content, &post.PictureContent, &post.CreatedAt)
		if err != nil{
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}