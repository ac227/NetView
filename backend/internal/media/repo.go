package media

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netview/netview/internal/db"
)

type Item struct {
	ID            int64     `json:"id"`
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	SourceURL     string    `json:"source_url"`
	LocalPath     string    `json:"local_path"`
	ThumbnailPath string    `json:"thumbnail_path"`
	MimeType      string    `json:"mime_type"`
	Size          int64     `json:"size"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	Duration      int       `json:"duration"`
	Status        string    `json:"status"`
	Favorite      bool      `json:"favorite"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Tags          []string  `json:"tags"`
	Categories    []int64   `json:"categories"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

type ListFilter struct {
	Keyword     string
	Type        string
	Status      string
	Favorite    *bool
	Tag         string
	Category    int64
	Sort        string
	Page        int
	PageSize    int
}

func (f *ListFilter) offset() int {
	return (f.Page - 1) * f.PageSize
}

func (r *Repo) Create(ctx context.Context, it *Item) (*Item, error) {
	if it.Status == "" {
		it.Status = "ready"
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO items (type, title, description, source_url, local_path, thumbnail_path,
			mime_type, size, width, height, duration, status, favorite)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, created_at, updated_at`,
		it.Type, it.Title, it.Description, it.SourceURL, it.LocalPath, it.ThumbnailPath,
		it.MimeType, it.Size, it.Width, it.Height, it.Duration, it.Status, it.Favorite,
	).Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}
	if err := r.setTags(ctx, it.ID, it.Tags); err != nil {
		return nil, err
	}
	if err := r.SetCategories(ctx, it.ID, it.Categories); err != nil {
		return nil, err
	}
	return r.Get(ctx, it.ID)
}

func (r *Repo) Get(ctx context.Context, id int64) (*Item, error) {
	it, err := r.scanItem(ctx, r.pool.QueryRow(ctx, `
		SELECT id, type, title, description, source_url, local_path, thumbnail_path,
			mime_type, size, width, height, duration, status, favorite, created_at, updated_at
		FROM items WHERE id=$1`, id))
	if err != nil {
		return nil, err
	}
	tags, err := r.tagsFor(ctx, id)
	if err != nil {
		return nil, err
	}
	cats, err := r.categoriesFor(ctx, id)
	if err != nil {
		return nil, err
	}
	it.Tags = tags
	it.Categories = cats
	return it, nil
}

func (r *Repo) Update(ctx context.Context, it *Item) (*Item, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE items SET title=$1, description=$2, source_url=$3, local_path=$4,
			thumbnail_path=$5, mime_type=$6, size=$7, width=$8, height=$9, duration=$10,
			status=$11, favorite=$12, updated_at=now()
		WHERE id=$13
		RETURNING updated_at`,
		it.Title, it.Description, it.SourceURL, it.LocalPath, it.ThumbnailPath,
		it.MimeType, it.Size, it.Width, it.Height, it.Duration, it.Status, it.Favorite, it.ID,
	).Scan(&it.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}
	if it.Tags != nil {
		if err := r.setTags(ctx, it.ID, it.Tags); err != nil {
			return nil, err
		}
	}
	if it.Categories != nil {
		if err := r.SetCategories(ctx, it.ID, it.Categories); err != nil {
			return nil, err
		}
	}
	return r.Get(ctx, it.ID)
}

func (r *Repo) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM items WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (r *Repo) SetFavorite(ctx context.Context, id int64, fav bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE items SET favorite=$2, updated_at=now() WHERE id=$1`, id, fav)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (r *Repo) List(ctx context.Context, f ListFilter) ([]*Item, int, error) {
	var conds []string
	var args []interface{}
	argc := 1
	if f.Keyword != "" {
		conds = append(conds, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argc, argc))
		args = append(args, "%"+f.Keyword+"%")
		argc++
	}
	if f.Type != "" {
		conds = append(conds, fmt.Sprintf("type=$%d", argc))
		args = append(args, f.Type)
		argc++
	}
	if f.Status != "" {
		conds = append(conds, fmt.Sprintf("status=$%d", argc))
		args = append(args, f.Status)
		argc++
	}
	if f.Favorite != nil {
		conds = append(conds, fmt.Sprintf("favorite=$%d", argc))
		args = append(args, *f.Favorite)
		argc++
	}
	if f.Tag != "" {
		conds = append(conds, `EXISTS (SELECT 1 FROM item_tags it2 JOIN tags t2 ON t2.id=it2.tag_id WHERE it2.item_id=items.id AND t2.name=$`+itoa(argc)+`)`)
		args = append(args, f.Tag)
		argc++
	}
	if f.Category > 0 {
		conds = append(conds, `EXISTS (SELECT 1 FROM item_categories ic2 WHERE ic2.item_id=items.id AND ic2.category_id=$`+itoa(argc)+`)`)
		args = append(args, f.Category)
		argc++
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM items"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sort := "created_at DESC"
	switch f.Sort {
	case "oldest":
		sort = "created_at ASC"
	case "name":
		sort = "title ASC"
	case "size":
		sort = "size DESC"
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}

	query := "SELECT id, type, title, description, source_url, local_path, thumbnail_path, " +
		"mime_type, size, width, height, duration, status, favorite, created_at, updated_at " +
		"FROM items" + where + " ORDER BY " + sort +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", argc, argc+1)
	args = append(args, f.PageSize, f.offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		it, err := scanItemRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	if items == nil {
		items = []*Item{}
	}
	if len(items) > 0 {
		if err := r.populateTagsAndCategories(ctx, items); err != nil {
			return nil, 0, err
		}
	}
	return items, total, nil
}

// populateTagsAndCategories fills tags/categories for a page of items in batch (avoids N+1).
func (r *Repo) populateTagsAndCategories(ctx context.Context, items []*Item) error {
	ids := make([]int64, len(items))
	byID := make(map[int64]*Item, len(items))
	for i, it := range items {
		ids[i] = it.ID
		byID[it.ID] = it
		it.Tags = []string{}
		it.Categories = []int64{}
	}

	rows, err := r.pool.Query(ctx, `
		SELECT it.item_id, t.name
		FROM item_tags it JOIN tags t ON t.id = it.tag_id
		WHERE it.item_id = ANY($1) ORDER BY t.name`, ids)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return err
		}
		if it, ok := byID[id]; ok {
			it.Tags = append(it.Tags, name)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	crows, err := r.pool.Query(ctx, `
		SELECT item_id, category_id FROM item_categories WHERE item_id = ANY($1)`, ids)
	if err != nil {
		return err
	}
	for crows.Next() {
		var id, cat int64
		if err := crows.Scan(&id, &cat); err != nil {
			crows.Close()
			return err
		}
		if it, ok := byID[id]; ok {
			it.Categories = append(it.Categories, cat)
		}
	}
	crows.Close()
	return crows.Err()
}

func (r *Repo) setTags(ctx context.Context, itemID int64, tags []string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM item_tags WHERE item_id=$1`, itemID); err != nil {
		return err
	}
	for _, name := range tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var tagID int64
		err := r.pool.QueryRow(ctx, `
			INSERT INTO tags (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
			RETURNING id`, name).Scan(&tagID)
		if err != nil {
			return err
		}
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO item_tags (item_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			itemID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) SetCategories(ctx context.Context, itemID int64, cats []int64) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM item_categories WHERE item_id=$1`, itemID); err != nil {
		return err
	}
	for _, cid := range cats {
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO item_categories (item_id, category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			itemID, cid); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) tagsFor(ctx context.Context, itemID int64) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.name FROM item_tags it JOIN tags t ON t.id=it.tag_id WHERE it.item_id=$1 ORDER BY t.name`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, nil
}

func (r *Repo) categoriesFor(ctx context.Context, itemID int64) ([]int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT category_id FROM item_categories WHERE item_id=$1`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		cats = append(cats, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if cats == nil {
		cats = []int64{}
	}
	return cats, nil
}

func (r *Repo) ListTags(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, nil
}

func (r *Repo) SetItemStatus(ctx context.Context, id int64, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE items SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	return err
}

// ApplyDownloadResult updates only download-related fields so that concurrent
// metadata edits (title/tags/description) are never overwritten.
func (r *Repo) ApplyDownloadResult(ctx context.Context, it *Item) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE items SET
			status = $2, local_path = $3, thumbnail_path = $4,
			mime_type = $5, size = $6, width = $7, height = $8, duration = $9,
			type = CASE WHEN $10 = '' THEN type ELSE $10 END,
			updated_at = now()
		WHERE id = $1`,
		it.ID, it.Status, it.LocalPath, it.ThumbnailPath,
		it.MimeType, it.Size, it.Width, it.Height, it.Duration, it.Type)
	return err
}

func (r *Repo) scanItem(ctx context.Context, row pgx.Row) (*Item, error) {
	it, err := scanItemRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	return it, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanItemRow(row rowScanner) (*Item, error) {
	var it Item
	err := row.Scan(&it.ID, &it.Type, &it.Title, &it.Description, &it.SourceURL,
		&it.LocalPath, &it.ThumbnailPath, &it.MimeType, &it.Size, &it.Width, &it.Height,
		&it.Duration, &it.Status, &it.Favorite, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
