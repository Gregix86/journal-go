-- Historique des versions des entrees (articles/recettes), facon Git :
-- chaque modification de contenu cree un nouveau snapshot avec un message.

CREATE TABLE IF NOT EXISTS entry_versions (
    id                  SERIAL PRIMARY KEY,
    entry_id            INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    version_number      INTEGER NOT NULL,
    title               VARCHAR(255) NOT NULL,
    excerpt             VARCHAR(500) NOT NULL DEFAULT '',
    content_markdown    TEXT NOT NULL DEFAULT '',
    entry_type          VARCHAR(20) NOT NULL DEFAULT 'article',
    category_id         INTEGER NOT NULL REFERENCES categories(id),
    recipe_servings     INTEGER,
    recipe_prep_minutes INTEGER,
    recipe_cook_minutes INTEGER,
    recipe_ingredients  JSONB,
    recipe_steps        JSONB,
    message             VARCHAR(500) NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (entry_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_entry_versions_entry ON entry_versions(entry_id, version_number DESC);
