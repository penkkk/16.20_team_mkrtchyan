BEGIN;

DROP TABLE IF EXISTS auth.external_identities;
DROP TABLE IF EXISTS auth.refresh_sessions;
DROP TABLE IF EXISTS auth.user_passwords;
DROP TABLE IF EXISTS auth.users;

DROP SCHEMA IF EXISTS auth;

COMMIT;
