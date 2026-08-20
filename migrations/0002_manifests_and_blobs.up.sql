CREATE TABLE manifests (
    vault_id   UUID        PRIMARY KEY REFERENCES vaults (id) ON DELETE CASCADE,
    generation BIGINT      NOT NULL,
    ciphertext BYTEA       NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE blobs (
    vault_id   UUID        NOT NULL REFERENCES vaults (id) ON DELETE CASCADE,
    id         TEXT        NOT NULL,
    size_bytes BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (vault_id, id)
);
