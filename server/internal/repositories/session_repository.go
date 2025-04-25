package repositories

import (
    "context"
    "time"
    
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    
    "black-lotus/internal/models"
)

type SessionRepository struct {
    db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
    return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, duration time.Duration) (*models.Session, error) {
    var session models.Session
    
    expiresAt := time.Now().Add(duration)
    
    err := r.db.QueryRow(ctx, `
        INSERT INTO sessions (user_id, expires_at)
        VALUES ($1, $2)
        RETURNING id, user_id, expires_at, created_at
    `, userID, expiresAt).Scan(
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

func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
    var session models.Session
    
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

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
    _, err := r.db.Exec(ctx, `
        DELETE FROM sessions
        WHERE id = $1
    `, sessionID)
    
    return err
}

func (r *SessionRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
    _, err := r.db.Exec(ctx, `
        DELETE FROM sessions
        WHERE user_id = $1
    `, userID)
    
    return err
}