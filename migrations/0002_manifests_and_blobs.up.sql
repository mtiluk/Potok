CREATE TABLE manifests (
    vault_id   TEXT      PRIMARY KEY REFERENCES vaults (id) ON DELETE CASCADE,
    generation INTEGER   NOT NULL,
    ciphertext BLOB      NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE blobs (
    vault_id   TEXT      NOT NULL REFERENCES vaults (id) ON DELETE CASCADE,
    id         TEXT      NOT NULL,
    size_bytes INTEGER   NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (vault_id, id)
);
