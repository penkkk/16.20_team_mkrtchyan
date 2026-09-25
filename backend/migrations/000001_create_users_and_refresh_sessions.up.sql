BEGIN;

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    tg_username TEXT UNIQUE,
    name TEXT,
    surname TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth.user_passwords (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS auth.refresh_sessions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    user_agent TEXT,
    ip INET,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth.external_identities (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_subject TEXT NOT NULL,
    provider_username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_subject),
    UNIQUE (user_id, provider)
);

CREATE INDEX IF NOT EXISTS external_identities_user_id_idx
    ON auth.external_identities(user_id);

CREATE INDEX refresh_sessions_token_hash_idx ON auth.refresh_sessions(token_hash);

CREATE UNIQUE INDEX IF NOT EXISTS user_passwords_user_id_uidx
    ON auth.user_passwords(user_id);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uidx
    ON auth.users(LOWER(email));

CREATE UNIQUE INDEX IF NOT EXISTS refresh_sessions_token_hash_uidx
    ON auth.refresh_sessions(token_hash);

CREATE INDEX IF NOT EXISTS refresh_sessions_user_id_idx
    ON auth.refresh_sessions(user_id);

CREATE INDEX IF NOT EXISTS refresh_sessions_active_idx
    ON auth.refresh_sessions(token_hash, expires_at)
    WHERE revoked_at IS NULL;

COMMIT;
