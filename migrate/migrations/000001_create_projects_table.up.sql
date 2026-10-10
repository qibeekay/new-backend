CREATE TABLE IF NOT EXISTS projects (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    tagline     TEXT NOT NULL,
    year        TEXT NOT NULL,
    category    TEXT NOT NULL CHECK (category IN ('creative','functional','systems')),
    status      TEXT NOT NULL CHECK (status IN ('completed','ongoing','ideating','paused')),
    featured    BOOLEAN NOT NULL DEFAULT FALSE,
    image       TEXT NOT NULL DEFAULT '',
    image_alt   TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT '',
    stack       TEXT[] NOT NULL DEFAULT '{}',
    metrics     JSONB NOT NULL DEFAULT '[]',
    live_url    TEXT NOT NULL DEFAULT '',
    code_url    TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS projects_created_idx ON projects (created_at DESC);
