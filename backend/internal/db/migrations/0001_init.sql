CREATE TABLE IF NOT EXISTS settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS items (
    id              BIGSERIAL PRIMARY KEY,
    type            TEXT NOT NULL CHECK (type IN ('image', 'video')),
    title           TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    source_url      TEXT NOT NULL DEFAULT '',
    local_path      TEXT NOT NULL DEFAULT '',
    thumbnail_path  TEXT NOT NULL DEFAULT '',
    mime_type       TEXT NOT NULL DEFAULT '',
    size            BIGINT NOT NULL DEFAULT 0,
    width           INTEGER NOT NULL DEFAULT 0,
    height          INTEGER NOT NULL DEFAULT 0,
    duration        INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'ready'
                    CHECK (status IN ('ready', 'downloading', 'downloaded', 'failed')),
    favorite        BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_items_created_at ON items (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_type ON items (type);
CREATE INDEX IF NOT EXISTS idx_items_status ON items (status);
CREATE INDEX IF NOT EXISTS idx_items_favorite ON items (favorite);

CREATE TABLE IF NOT EXISTS tags (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS item_tags (
    item_id BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    tag_id  BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_item_tags_tag ON item_tags (tag_id);

CREATE TABLE IF NOT EXISTS categories (
    id        BIGSERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    parent_id BIGINT REFERENCES categories(id) ON DELETE CASCADE,
    sort      INTEGER NOT NULL DEFAULT 0,
    UNIQUE (parent_id, name)
);

CREATE TABLE IF NOT EXISTS item_categories (
    item_id       BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    category_id   BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, category_id)
);

CREATE TABLE IF NOT EXISTS download_jobs (
    id          BIGSERIAL PRIMARY KEY,
    item_id     BIGINT REFERENCES items(id) ON DELETE CASCADE,
    url         TEXT NOT NULL DEFAULT '',
    adapter     TEXT NOT NULL DEFAULT 'direct',
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'running', 'done', 'failed', 'cancelled')),
    progress    REAL NOT NULL DEFAULT 0,
    info        TEXT NOT NULL DEFAULT '',
    error       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_download_jobs_item ON download_jobs (item_id);
CREATE INDEX IF NOT EXISTS idx_download_jobs_status ON download_jobs (status);
