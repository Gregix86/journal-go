-- Schema initial du carnet.
-- Applique automatiquement au demarrage du serveur (voir internal/db migrate.go).

CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    username        VARCHAR(80) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS categories (
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(80) UNIQUE NOT NULL,
    name        VARCHAR(120) NOT NULL,
    accent      VARCHAR(20) NOT NULL DEFAULT '#1F5C4A',
    description VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS entries (
    id               SERIAL PRIMARY KEY,
    category_id      INTEGER NOT NULL REFERENCES categories(id),
    entry_type       VARCHAR(20) NOT NULL DEFAULT 'article',
    title            VARCHAR(255) NOT NULL,
    slug             VARCHAR(255) UNIQUE NOT NULL,
    excerpt          VARCHAR(500) NOT NULL DEFAULT '',
    content_markdown TEXT NOT NULL DEFAULT '',
    cover_media_id   INTEGER,
    published        BOOLEAN NOT NULL DEFAULT true,
    is_private       BOOLEAN NOT NULL DEFAULT false,
    entry_number     INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS media (
    id          SERIAL PRIMARY KEY,
    entry_id    INTEGER REFERENCES entries(id) ON DELETE CASCADE,
    filename    VARCHAR(255) NOT NULL,
    filepath    VARCHAR(500) NOT NULL,
    kind        VARCHAR(20) NOT NULL DEFAULT 'image',
    mime_type   VARCHAR(100) NOT NULL DEFAULT '',
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE entries
    DROP CONSTRAINT IF EXISTS entries_cover_media_id_fkey;
ALTER TABLE entries
    ADD CONSTRAINT entries_cover_media_id_fkey
    FOREIGN KEY (cover_media_id) REFERENCES media(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS recipe_details (
    id            SERIAL PRIMARY KEY,
    entry_id      INTEGER UNIQUE NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    servings      INTEGER NOT NULL DEFAULT 4,
    prep_minutes  INTEGER NOT NULL DEFAULT 0,
    cook_minutes  INTEGER NOT NULL DEFAULT 0,
    ingredients   JSONB NOT NULL DEFAULT '[]',
    steps         JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS devices (
    id         SERIAL PRIMARY KEY,
    slug       VARCHAR(80) UNIQUE NOT NULL,
    name       VARCHAR(120) NOT NULL,
    api_key    VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sensor_readings (
    id          SERIAL PRIMARY KEY,
    device_id   INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric      VARCHAR(50) NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        VARCHAR(20) NOT NULL DEFAULT '',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_entries_category ON entries(category_id);
CREATE INDEX IF NOT EXISTS idx_entries_published_created ON entries(published, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_readings_device_metric_time ON sensor_readings(device_id, metric, recorded_at DESC);
