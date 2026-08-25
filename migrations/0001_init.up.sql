CREATE TABLE users (
    id            TEXT      PRIMARY KEY,
    email         TEXT      NOT NULL UNIQUE,
    password_hash TEXT      NOT NULL,
    is_admin      INTEGER   NOT NULL DEFAULT 0,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE api_keys (
    id           TEXT      PRIMARY KEY,
    user_id      TEXT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    label        TEXT      NOT NULL,
    key_hash     BLOB      NOT NULL UNIQUE,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP
);

CREATE INDEX api_keys_user_id_idx ON api_keys (user_id);

CREATE TABLE vaults (
    id          TEXT      PRIMARY KEY,
    user_id     TEXT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT      NOT NULL,
    wrapped_key BLOB      NOT NULL,
    generation  INTEGER   NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, name)
);

CREATE INDEX vaults_user_id_idx ON vaults (user_id);
