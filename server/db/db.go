package db

import (
	"context"
	"fmt"
	"os"
	"time"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// Initialize sets up the database connection
func Initialize() error {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	var err error
	DB, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}

	// Verify connection
	if err := DB.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %v", err)
	}

	// Initialize schema
	if err := initSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	return nil
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}

// initSchema creates database tables if they don't exist
func initSchema() error {
    _, err := DB.Exec(context.Background(), `
        -- Enable UUID extension if not already enabled
        CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
        
        -- Users table
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(100) NOT NULL,
            email VARCHAR(100) UNIQUE NOT NULL,
            hashed_password VARCHAR(255) DEFAULT NULL,
            email_verified BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
        );
        
        -- OAuth accounts table
        CREATE TABLE IF NOT EXISTS oauth_accounts (
            provider_id VARCHAR(100) NOT NULL,
            provider_user_id VARCHAR(100) NOT NULL,
            user_id UUID NOT NULL,
            access_token TEXT NOT NULL,
            refresh_token TEXT DEFAULT NULL,
            expires_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (provider_id, provider_user_id),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        
        -- Sessions table
        CREATE TABLE IF NOT EXISTS sessions (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL,
            expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        
        -- Email verification table
        CREATE TABLE IF NOT EXISTS email_verifications (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            code VARCHAR(100) NOT NULL,
            user_id UUID NOT NULL UNIQUE,
            expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours'),
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        
        -- Create indexes for better performance
        CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id);
        CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
        CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
        CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at);
    `)
    
    return err
}

// CleanupExpiredRecords removes all expired sessions and verification codes
func CleanupExpiredRecords(ctx context.Context) (int64, error) {
	// Delete expired sessions
	sessionResult, err := DB.Exec(ctx, `
		DELETE FROM sessions WHERE expires_at < NOW()
	`)
	if err != nil {
		return 0, err
	}
	
	// Delete expired email verifications
	verificationResult, err := DB.Exec(ctx, `
		DELETE FROM email_verifications WHERE expires_at < NOW()
	`)
	if err != nil {
		return 0, err
	}
	
	// Return total number of deleted records
	sessionCount := sessionResult.RowsAffected()
	verificationCount := verificationResult.RowsAffected()
	
	return sessionCount + verificationCount, nil
}

// StartCleanupJob starts a background goroutine that periodically cleans up expired records
func StartCleanupJob(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				count, err := CleanupExpiredRecords(context.Background())
				if err != nil {
					log.Printf("Error cleaning up expired records: %v", err)
				} else if count > 0 {
					log.Printf("Cleaned up %d expired records", count)
				}
			}
		}
	}()
}