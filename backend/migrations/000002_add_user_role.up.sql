BEGIN;

ALTER TABLE auth.users
    ADD COLUMN role TEXT NOT NULL DEFAULT 'user'
    CHECK (role IN ('user', 'admin'));

COMMIT;
