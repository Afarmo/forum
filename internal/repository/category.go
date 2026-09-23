package repository

import (
	"context"
	"database/sql"

	"github.com/Afarmo/forum/internal/models"
)


type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	query := `SELECT id, name FROM categories ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *CategoryRepository) GetCategoryByUser(ctx context.Context, userId int) ([]models.Category, error) {
	query := `SELECT CATEGORIES.id, CATEGORIES.name 
	FROM CATEGORIES 
	JOIN POST_CATEGORIES 
		ON CATEGORIES.id = POST_CATEGORIES.category_id
	JOIN POSTS
		ON POST_CATEGORIES.post_id = POSTS.id
	WHERE POSTS.user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []models.Category

	for rows.Next() {
		var category models.Category
		err := rows.Scan( &category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}
