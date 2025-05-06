package repositories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"black-lotus/internal/models"
)

// SessionRepository handles database operations for sessions
type SessionRepository struct {
	db *pgxpool.Pool // Database connection pool
}

type SessionRepositoryInterface interface {
	CreateSession(ctx context.Context, userID uuid.UUID, accessDuration, refreshDuration time.Duration) (*models.Session, error)
	GetSessionByAccessToken(ctx context.Context, token string) (*models.Session, error)
	GetSessionByRefreshToken(ctx context.Context, token string) (*models.Session, error)
	RefreshAccessToken(ctx context.Context, sessionID uuid.UUID) (*models.Session, error)
	DeleteSessionByAccessToken(ctx context.Context, token string) error
	DeleteSessionByRefreshToken(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
}

// NewSessionRepository creates a new repository with the given database connection
func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession stores a new session with both access and refresh tokens
func (r *SessionRepository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) (*models.Session, error) {
	session := new(models.Session)

	// Generate access token
	accessTokenBytes := make([]byte, 32)
	if _, err := rand.Read(accessTokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	accessToken := base64.StdEncoding.EncodeToString(accessTokenBytes)
	accessHash := sha256.Sum256([]byte(accessToken))
	accessTokenHash := hex.EncodeToString(accessHash[:])

	// Generate refresh token
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken := base64.StdEncoding.EncodeToString(refreshTokenBytes)
	refreshHash := sha256.Sum256([]byte(refreshToken))
	refreshTokenHash := hex.EncodeToString(refreshHash[:])

	// Set expiration times
	accessExpiry := time.Now().Add(accessDuration)
	refreshExpiry := time.Now().Add(refreshDuration)

	// Insert into database
	err := r.db.QueryRow(ctx, `
        INSERT INTO sessions (user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, user_id, access_expires_at, refresh_expires_at, created_at
    `, userID, accessTokenHash, refreshTokenHash, accessExpiry, refreshExpiry).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessExpiry,
		&session.RefreshExpiry,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert session: %w", err)
	}

	// Save tokens in the session object
	session.AccessToken = accessToken
	session.RefreshToken = refreshToken

	return session, nil
}

// GetSessionByAccessToken retrieves a session using an access token
func (r *SessionRepository) GetSessionByAccessToken(ctx context.Context, token string) (*models.Session, error) {
	session := new(models.Session)

	// Hash the token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Query by token hash
	err := r.db.QueryRow(ctx, `
        SELECT id, user_id, access_expires_at, refresh_expires_at, created_at
        FROM sessions
        WHERE access_token_hash = $1 AND access_expires_at > NOW()
    `, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessExpiry,
		&session.RefreshExpiry,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSessionByRefreshToken retrieves a session using a refresh token
func (r *SessionRepository) GetSessionByRefreshToken(ctx context.Context, token string) (*models.Session, error) {
	session := new(models.Session)

	// Hash the token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Query by token hash
	err := r.db.QueryRow(ctx, `
        SELECT id, user_id, access_expires_at, refresh_expires_at, created_at
        FROM sessions
        WHERE refresh_token_hash = $1 AND refresh_expires_at > NOW()
    `, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessExpiry,
		&session.RefreshExpiry,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

// RefreshAccessToken generates a new access token for a session
func (r *SessionRepository) RefreshAccessToken(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	session := new(models.Session)

	// Generate new access token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	accessToken := base64.StdEncoding.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(accessToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Set new expiration time (1 hour from now)
	accessExpiry := time.Now().Add(1 * time.Hour)

	// Update in database
	err := r.db.QueryRow(ctx, `
        UPDATE sessions
        SET access_token_hash = $1, access_expires_at = $2
        WHERE id = $3
        RETURNING id, user_id, access_expires_at, refresh_expires_at, created_at
    `, tokenHash, accessExpiry, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessExpiry,
		&session.RefreshExpiry,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Set the new access token
	session.AccessToken = accessToken

	return session, nil
}

// DeleteSessionByAccessToken removes a session using its access token
func (r *SessionRepository) DeleteSessionByAccessToken(ctx context.Context, token string) error {
	// Hash the token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Delete using the token hash
	_, err := r.db.Exec(ctx, `
        DELETE FROM sessions
        WHERE access_token_hash = $1
    `, tokenHash)

	return err
}

// DeleteSessionByRefreshToken removes a session using its refresh token
func (r *SessionRepository) DeleteSessionByRefreshToken(ctx context.Context, token string) error {
	// Hash the token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Delete using the token hash
	_, err := r.db.Exec(ctx, `
        DELETE FROM sessions
        WHERE refresh_token_hash = $1
    `, tokenHash)

	return err
}

// DeleteUserSessions removes all sessions for a specific user
func (r *SessionRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM sessions
		WHERE user_id = $1
	`, userID)

	return err
}
