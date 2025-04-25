package repositories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

// NewSessionRepository creates a new repository with the given database connection
func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}


// CreateSession stores a new session in the database
// Generates a secure token hash to prevent session hijacking
func (r *SessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, duration time.Duration) (*models.Session, error) {
	var session models.Session
	
	// Generate a random token for the session
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	// Hash the token for storage
	hash := sha256.Sum256(tokenBytes)
	tokenHash := hex.EncodeToString(hash[:])
	
	expiresAt := time.Now().Add(duration)
	
	// Use the token_hash in the SQL query
	err := r.db.QueryRow(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, expires_at, created_at
	`, userID, tokenHash, expiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to insert session: %w", err)
	}
	
	return &session, nil
}

// GetSessionByID retrieves a valid (non-expired) session by its ID
func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	var session models.Session
	
	// Query that only returns non-expired sessions
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = $1 AND expires_at > NOW()
	`, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &session, nil
}

// DeleteSession removes a session from the database
func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM sessions
		WHERE id = $1
	`, sessionID)
	
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