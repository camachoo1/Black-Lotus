package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"black-lotus/internal/domain/models"
	"black-lotus/internal/features/auth/oauth/github"
	"black-lotus/internal/features/auth/oauth/google"
)

// OAuthRepository handles database operations for OAuth accounts
type OAuthRepository struct {
	db *pgxpool.Pool
}

var (
	_ github.OAuthRepository = (*OAuthRepository)(nil)
	_ google.OAuthRepository = (*OAuthRepository)(nil)
)

// NewOAuthRepository creates a new repository with database connection
func NewOAuthRepository(db *pgxpool.Pool) *OAuthRepository {
	return &OAuthRepository{db: db}
}

// CreateOAuthAccount creates or updates an OAuth account
func (r *OAuthRepository) CreateOAuthAccount(ctx context.Context, account models.OAuthAccount) error {
	// Use upsert to handle both new accounts and reconnecting existing ones
	_, err := r.db.Exec(ctx, `
		INSERT INTO oauth_accounts (
			provider_id, provider_user_id, user_id, 
			access_token, refresh_token, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (provider_id, provider_user_id) 
		DO UPDATE SET 
			user_id = $3,
			access_token = $4, 
			refresh_token = $5, 
			expires_at = $6,
			updated_at = CURRENT_TIMESTAMP
	`, account.ProviderID, account.ProviderUserID, account.UserID,
		account.AccessToken, account.RefreshToken, account.ExpiresAt)

	return err
}

// GetOAuthAccount gets the OAuth account by provider and the provider user ID
func (r *OAuthRepository) GetOAuthAccount(ctx context.Context, providerID, providerUserID string) (*models.OAuthAccount, error) {
	account := new(models.OAuthAccount)

	err := r.db.QueryRow(ctx, `
		SELECT provider_id, provider_user_id, user_id, 
			access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE provider_id = $1 AND provider_user_id = $2
	`, providerID, providerUserID).Scan(
		&account.ProviderID,
		&account.ProviderUserID,
		&account.UserID,
		&account.AccessToken,
		&account.RefreshToken,
		&account.ExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetUserOAuthAccounts gets all OAuth accounts for a specific user
func (r *OAuthRepository) GetUserOAuthAccounts(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	rows, err := r.db.Query(ctx, `
		SELECT provider_id, provider_user_id, user_id, 
			access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*models.OAuthAccount{}
	for rows.Next() {
		account := &models.OAuthAccount{}
		err := rows.Scan(
			&account.ProviderID,
			&account.ProviderUserID,
			&account.UserID,
			&account.AccessToken,
			&account.RefreshToken,
			&account.ExpiresAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
