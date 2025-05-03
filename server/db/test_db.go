package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var TestDB *pgxpool.Pool

// InitializeTestDB sets up the test database connection
func InitializeTestDB() error {
	// Use environment variables or default to standard test values
	testUser := os.Getenv("TEST_DB_USER")
	if testUser == "" {
		testUser = "postgres"
	}

	testPassword := os.Getenv("TEST_DB_PASSWORD")
	if testPassword == "" {
		testPassword = "postgres"
	}

	testHost := os.Getenv("TEST_DB_HOST")
	if testHost == "" {
		testHost = "localhost"
	}

	testPort := os.Getenv("TEST_DB_PORT")
	if testPort == "" {
		testPort = "5432"
	}

	testDBName := os.Getenv("TEST_DB_NAME")
	if testDBName == "" {
		testDBName = "black_lotus_test"
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		testUser, testPassword, testHost, testPort, testDBName)

	var err error
	TestDB, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("unable to connect to test database: %v", err)
	}

	// Verify connection
	if err := TestDB.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping test database: %v", err)
	}

	// Initialize schema for test database
	if err := initTestSchema(); err != nil {
		return fmt.Errorf("failed to initialize test schema: %v", err)
	}

	return nil
}

// CloseTestDB closes the test database connection
func CloseTestDB() {
	if TestDB != nil {
		TestDB.Close()
	}
}

// initTestSchema creates necessary tables for testing
func initTestSchema() error {
	_, err := TestDB.Exec(context.Background(), `
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
						CONSTRAINT email_format_check 
        		CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,4}$')
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
        
        -- Sessions table - updated for access & refresh tokens
        CREATE TABLE IF NOT EXISTS sessions (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL,
            access_token_hash VARCHAR(255),
            refresh_token_hash VARCHAR(255),
            access_expires_at TIMESTAMP WITH TIME ZONE,
            refresh_expires_at TIMESTAMP WITH TIME ZONE,
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
        
        -- Trips table
        CREATE TABLE IF NOT EXISTS trips (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL,
            name VARCHAR(100) NOT NULL,
            description TEXT,
            start_date TIMESTAMP WITH TIME ZONE NOT NULL,
            end_date TIMESTAMP WITH TIME ZONE NOT NULL,
            destination VARCHAR(255) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

				-- Create indexes for better performance
        CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id);
        CREATE INDEX IF NOT EXISTS idx_sessions_access_expires_at ON sessions(access_expires_at);
        CREATE INDEX IF NOT EXISTS idx_sessions_refresh_expires_at ON sessions(refresh_expires_at);
        CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
        CREATE INDEX IF NOT EXISTS idx_sessions_access_token_hash ON sessions(access_token_hash);
        CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token_hash ON sessions(refresh_token_hash);
        CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at);
    `)

	return err
}

// CleanTestTables truncates all test tables to ensure clean state
func CleanTestTables(ctx context.Context) error {
	_, err := TestDB.Exec(ctx, `
        TRUNCATE TABLE trips CASCADE;
        TRUNCATE TABLE users CASCADE;
    `)
	return err
}
