package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Vault struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func FetchUserVaults(userID int) ([]Vault, error) {
	rows, err := DB.Query("SELECT id, user_id, name, created_at, updated_at FROM vaults WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vaults := []Vault{}
	for rows.Next() {
		var v Vault
		if err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		vaults = append(vaults, v)
	}

	return vaults, rows.Err()
}

func FetchUserVaultByName(apiKey, vaultName string) (*Vault, error) {
	user, err := AuthenticateByKey(apiKey)
	if err != nil {
		return nil, errors.New("user not found")
	}

	var v Vault
	err = DB.QueryRow("SELECT id, user_id, name, created_at, updated_at FROM vaults WHERE user_id = ? AND name = ?", user.Id, vaultName).Scan(&v.ID, &v.UserID, &v.Name, &v.CreatedAt, &v.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("vault not found")
	} else if err != nil {
		return nil, err
	}

	return &v, nil
}

func CreateVault(userID int, vaultName string) (*Vault, error) {
	now := time.Now().UTC()
	res, err := DB.Exec(
		`INSERT INTO vaults (user_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		userID, vaultName, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert vault: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &Vault{
		ID:        int(id),
		UserID:    userID,
		Name:      vaultName,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
