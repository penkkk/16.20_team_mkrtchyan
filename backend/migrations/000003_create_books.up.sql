CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT NOT NULL,
    access_level TEXT NOT NULL CHECK (access_level IN ('free', 'subscription')),
    level TEXT NOT NULL CHECK (level IN ('beginner', 'intermediate', 'advanced')),
    topic TEXT NOT NULL,
    book_key TEXT NOT NULL UNIQUE,
    cover_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
