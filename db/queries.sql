-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (username, hashed_password) VALUES ($1, $2)
ON CONFLICT (username) DO NOTHING
RETURNING *;

-- name: ListCategories :many
SELECT * FROM categories ORDER BY id;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- name: CreateCategoryIfMissing :exec
INSERT INTO categories (slug, name, accent, description)
VALUES ($1, $2, $3, $4)
ON CONFLICT (slug) DO NOTHING;

-- name: CreateCategory :one
INSERT INTO categories (slug, name, accent, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateCategory :exec
UPDATE categories SET name = $2, description = $3, accent = $4 WHERE id = $1;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: CountEntriesInCategory :one
SELECT count(*) FROM entries WHERE category_id = $1;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: ListLatestPublicEntries :many
SELECT e.*, c.slug AS category_slug, c.name AS category_name, c.accent AS category_accent
FROM entries e
JOIN categories c ON c.id = e.category_id
WHERE e.published = true AND e.is_private = false
ORDER BY e.created_at DESC
LIMIT $1;

-- name: ListLatestAnyVisibilityEntries :many
SELECT e.*, c.slug AS category_slug, c.name AS category_name, c.accent AS category_accent
FROM entries e
JOIN categories c ON c.id = e.category_id
WHERE e.published = true
ORDER BY e.created_at DESC
LIMIT $1;

-- name: ListEntriesByCategoryPublic :many
SELECT e.*, c.slug AS category_slug, c.name AS category_name, c.accent AS category_accent
FROM entries e
JOIN categories c ON c.id = e.category_id
WHERE c.slug = $1 AND e.published = true AND e.is_private = false
ORDER BY e.created_at DESC;

-- name: ListEntriesByCategoryAnyVisibility :many
SELECT e.*, c.slug AS category_slug, c.name AS category_name, c.accent AS category_accent
FROM entries e
JOIN categories c ON c.id = e.category_id
WHERE c.slug = $1 AND e.published = true
ORDER BY e.created_at DESC;

-- name: ListAllEntries :many
SELECT e.*, c.slug AS category_slug, c.name AS category_name, c.accent AS category_accent
FROM entries e
JOIN categories c ON c.id = e.category_id
ORDER BY e.created_at DESC;

-- name: GetEntryBySlug :one
SELECT e.*, c.slug AS category_slug, c.name AS category_name, c.accent AS category_accent
FROM entries e
JOIN categories c ON c.id = e.category_id
WHERE e.slug = $1;

-- name: GetEntryByID :one
SELECT e.*, c.slug AS category_slug, c.name AS category_name, c.accent AS category_accent
FROM entries e
JOIN categories c ON c.id = e.category_id
WHERE e.id = $1;

-- name: CountEntries :one
SELECT count(*) FROM entries;

-- name: CreateEntry :one
INSERT INTO entries (category_id, entry_type, title, slug, excerpt, content_markdown, published, is_private, entry_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateEntry :exec
UPDATE entries
SET category_id = $2, entry_type = $3, title = $4, excerpt = $5,
    content_markdown = $6, published = $7, is_private = $8, updated_at = now()
WHERE id = $1;

-- name: SetEntryCover :exec
UPDATE entries SET cover_media_id = $2 WHERE id = $1;

-- name: DeleteEntry :exec
DELETE FROM entries WHERE id = $1;

-- name: EntrySlugExists :one
SELECT EXISTS(SELECT 1 FROM entries WHERE slug = $1);

-- name: CreateMedia :one
INSERT INTO media (entry_id, filename, filepath, kind, mime_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListMediaByEntry :many
SELECT * FROM media WHERE entry_id = $1 ORDER BY id;

-- name: DeleteMediaByEntry :exec
DELETE FROM media WHERE entry_id = $1;

-- name: GetRecipeDetail :one
SELECT * FROM recipe_details WHERE entry_id = $1;

-- name: UpsertRecipeDetail :exec
INSERT INTO recipe_details (entry_id, servings, prep_minutes, cook_minutes, ingredients, steps)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (entry_id) DO UPDATE
SET servings = $2, prep_minutes = $3, cook_minutes = $4, ingredients = $5, steps = $6;

-- name: DeleteRecipeDetail :exec
DELETE FROM recipe_details WHERE entry_id = $1;

-- name: CreateEntryVersion :one
INSERT INTO entry_versions (
    entry_id, version_number, title, excerpt, content_markdown, entry_type, category_id,
    recipe_servings, recipe_prep_minutes, recipe_cook_minutes, recipe_ingredients, recipe_steps, message
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: ListEntryVersions :many
SELECT * FROM entry_versions WHERE entry_id = $1 ORDER BY version_number DESC;

-- name: GetEntryVersion :one
SELECT * FROM entry_versions WHERE entry_id = $1 AND version_number = $2;

-- name: GetLatestVersionNumber :one
SELECT COALESCE(MAX(version_number), 0)::int FROM entry_versions WHERE entry_id = $1;
