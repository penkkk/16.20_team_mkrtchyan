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
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL
  AND expires_at > NOW()
RETURNING
    user_id,
    user_agent;

-- name: RevokeActiveRefreshSessionsByUserAgent :exec
UPDATE auth.refresh_sessions
SET revoked_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND user_agent IS NOT DISTINCT FROM sqlc.narg(user_agent)
  AND revoked_at IS NULL
  AND expires_at > NOW();

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

-- name: FindUserConflicts :many
SELECT field
FROM (
  SELECT 1 AS priority, 'username' AS field
  WHERE EXISTS (SELECT 1 FROM auth.users AS u WHERE u.username = sqlc.arg(username))
  UNION ALL
  SELECT 2 AS priority, 'email' AS field
  WHERE EXISTS (SELECT 1 FROM auth.users AS u WHERE LOWER(u.email) = LOWER(sqlc.arg(email)::text))
  UNION ALL
  SELECT 3 AS priority, 'tg_username' AS field
  WHERE sqlc.narg(tg_username)::text IS NOT NULL
    AND EXISTS (SELECT 1 FROM auth.users AS u WHERE u.tg_username = sqlc.narg(tg_username))
) AS conflicts
ORDER BY priority
;

-- name: GetUserByExternalIdentity :one
SELECT
    u.id,
    u.username,
    u.email,
    u.tg_username,
    u.name,
    u.surname,
    u.created_at,
    u.updated_at
FROM auth.external_identities AS ei
JOIN auth.users AS u ON u.id = ei.user_id
WHERE ei.provider = $1
  AND ei.provider_subject = $2
LIMIT 1;

-- name: CreateExternalIdentity :one
INSERT INTO auth.external_identities (
    user_id,
    provider,
    provider_subject,
    provider_username
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING
    id,
    user_id,
    provider,
    provider_subject,
    provider_username,
    created_at,
    updated_at;

-- name: GetUserPasswordByID :one
SELECT user_id, password_hash
FROM auth.user_passwords
WHERE user_id = sqlc.arg(user_id);
