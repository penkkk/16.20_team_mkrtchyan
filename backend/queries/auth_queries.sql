-- name: CreateUser :one
INSERT INTO auth.users (
    username,
    email,
    tg_username,
    name,
    surname
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING
    id,
    username,
    email,
    tg_username,
    name,
    surname,
    created_at,
    updated_at;

-- name: CreateUserPassword :one
INSERT INTO auth.user_passwords (
    user_id,
    password_hash
) VALUES (
    $1,
    $2
)
RETURNING
    id,
    user_id,
    password_hash;

-- name: CreateRefreshSession :one
INSERT INTO auth.refresh_sessions (
    user_id,
    token_hash,
    user_agent,
    ip,
    expires_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING
    id,
    user_id,
    token_hash,
    user_agent,
    ip,
    expires_at,
    revoked_at,
    created_at;

-- name: RevokeRefreshSession :one
UPDATE auth.refresh_sessions
SET revoked_at = NOW()
WHERE token_hash = $1
  AND revoked_at IS NULL
RETURNING
    id,
    user_id,
    token_hash,
    user_agent,
    ip,
    expires_at,
    revoked_at,
    created_at;

-- name: GetUserCredentialsByLogin :one
SELECT
    u.id,
    u.username,
    u.email,
    u.tg_username,
    u.name,
    u.surname,
    u.created_at,
    u.updated_at,
    up.password_hash
FROM auth.users AS u
JOIN auth.user_passwords AS up ON up.user_id = u.id
WHERE u.username = $1
   OR LOWER(u.email) = LOWER($1::text)
LIMIT 1;
