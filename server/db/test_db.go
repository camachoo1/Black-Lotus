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
        
        -- Users table (simplified for testing)
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(100) NOT NULL,
            email VARCHAR(100) UNIQUE NOT NULL,
            hashed_password VARCHAR(255) DEFAULT NULL,
            email_verified BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
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
