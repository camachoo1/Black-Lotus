package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var TestDB *pgxpool.Pool

const maxRetries = 3
const retryDelay = 2 * time.Second

// InitializeTestDB sets up the test database connection with retry logic
func InitializeTestDB() error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying database initialization (attempt %d/%d)...", attempt+1, maxRetries)
			time.Sleep(retryDelay)
		}

		err = initializeTestDBWithCleanup()
		if err == nil {
			return nil
		}

		log.Printf("Database initialization attempt failed: %v", err)
		// Make sure any partial connections are closed before retrying
		CloseTestDB()
	}

	return fmt.Errorf("failed to initialize test database after %d attempts: %v", maxRetries, err)
}

// initializeTestDBWithCleanup handles the actual initialization process
func initializeTestDBWithCleanup() error {
	// Connect to default postgres database first to recreate test database
	defaultConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres",
		getTestEnv("TEST_DB_USER", "postgres"),
		getTestEnv("TEST_DB_PASSWORD", "postgres"),
		getTestEnv("TEST_DB_HOST", "localhost"),
		getTestEnv("TEST_DB_PORT", "5432"))

	log.Printf("Connecting to default database: %s", defaultConnString)

	// Connect to the default postgres database
	defaultDB, err := pgxpool.New(context.Background(), defaultConnString)
	if err != nil {
		return fmt.Errorf("unable to connect to default postgres database: %v", err)
	}
	defer defaultDB.Close()

	// Check if our test database exists
	testDBName := getTestEnv("TEST_DB_NAME", "black_lotus_test")
	log.Printf("Checking if test database %s exists", testDBName)

	var exists bool
	err = defaultDB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		testDBName).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking if database exists: %v", err)
	}

	// Drop the test database if it exists and recreate it
	if exists {
		log.Printf("Dropping existing test database: %s", testDBName)
		// Close any existing connections to the database (improved version with timeout)
		_, err = defaultDB.Exec(context.Background(),
			fmt.Sprintf("SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid()", testDBName))
		if err != nil {
			log.Printf("Warning: Error terminating connections: %v", err)
			// Continue anyway, as the database might still be droppable
		}

		// Add a small delay to allow connections to close fully
		time.Sleep(500 * time.Millisecond)

		// Drop the database with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = defaultDB.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
		if err != nil {
			return fmt.Errorf("error dropping test database: %v", err)
		}
	}

	// Create the test database
	log.Printf("Creating test database: %s", testDBName)
	_, err = defaultDB.Exec(context.Background(),
		fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		return fmt.Errorf("error creating test database: %v", err)
	}
	log.Printf("Test database created successfully")

	// Now connect to the test database
	testConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		getTestEnv("TEST_DB_USER", "postgres"),
		getTestEnv("TEST_DB_PASSWORD", "postgres"),
		getTestEnv("TEST_DB_HOST", "localhost"),
		getTestEnv("TEST_DB_PORT", "5432"),
		testDBName)

	log.Printf("Connecting to test database: %s", testConnString)
	TestDB, err = pgxpool.New(context.Background(), testConnString)
	if err != nil {
		return fmt.Errorf("unable to connect to test database: %v", err)
	}

	// Verify connection
	if err := TestDB.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping test database: %v", err)
	}
	log.Printf("Successfully connected to test database")

	// Initialize schema
	log.Printf("Initializing test database schema")
	if err := initTestSchema(); err != nil {
		return fmt.Errorf("failed to initialize test schema: %v", err)
	}
	log.Printf("Schema initialized successfully")

	return nil
}

// CloseTestDB closes the test database connection
func CloseTestDB() {
	if TestDB != nil {
		TestDB.Close()
		TestDB = nil
	}
}

// Helper to get environment variable with default fallback
func getTestEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// initTestSchema creates necessary tables for testing
func initTestSchema() error {
	// Create UUID extension
	log.Printf("Creating UUID extension")
	_, err := TestDB.Exec(context.Background(),
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		return fmt.Errorf("failed to create UUID extension: %v", err)
	}

	// Create users table with fixed email validation constraint
	log.Printf("Creating users table")
	_, err = TestDB.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(100) NOT NULL,
            email VARCHAR(100) UNIQUE NOT NULL,
            hashed_password VARCHAR(255) DEFAULT NULL,
            email_verified BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT email_format_check 
            CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4}$')
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create trips table with location column
	log.Printf("Creating trips table with location column")
	_, err = TestDB.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS trips (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			start_date TIMESTAMP WITH TIME ZONE NOT NULL,
			end_date TIMESTAMP WITH TIME ZONE NOT NULL,
			location VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
  `)
	if err != nil {
		return fmt.Errorf("failed to create trips table: %v", err)
	}

	// Create oauth_accounts table
	log.Printf("Creating oauth_accounts table")
	_, err = TestDB.Exec(context.Background(), `
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
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create oauth_accounts table: %v", err)
	}

	// Create sessions table
	log.Printf("Creating sessions table")
	_, err = TestDB.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL,
			access_token_hash VARCHAR(255),
			refresh_token_hash VARCHAR(255),
			access_expires_at TIMESTAMP WITH TIME ZONE,
			refresh_expires_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %v", err)
	}

	// Create email_verifications table
	log.Printf("Creating email_verifications table")
	_, err = TestDB.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS email_verifications (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			code VARCHAR(100) NOT NULL,
			user_id UUID NOT NULL UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours'),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create email_verifications table: %v", err)
	}

	// Create all indexes
	log.Printf("Creating indexes for oauth_accounts")
	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id)")
	if err != nil {
		return fmt.Errorf("failed to create oauth_accounts index: %v", err)
	}

	log.Printf("Creating indexes for sessions")
	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_sessions_access_expires_at ON sessions(access_expires_at)")
	if err != nil {
		return fmt.Errorf("failed to create sessions access_expires_at index: %v", err)
	}

	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_sessions_refresh_expires_at ON sessions(refresh_expires_at)")
	if err != nil {
		return fmt.Errorf("failed to create sessions refresh_expires_at index: %v", err)
	}

	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)")
	if err != nil {
		return fmt.Errorf("failed to create sessions user_id index: %v", err)
	}

	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_sessions_access_token_hash ON sessions(access_token_hash)")
	if err != nil {
		return fmt.Errorf("failed to create sessions access_token_hash index: %v", err)
	}

	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token_hash ON sessions(refresh_token_hash)")
	if err != nil {
		return fmt.Errorf("failed to create sessions refresh_token_hash index: %v", err)
	}

	log.Printf("Creating indexes for email_verifications")
	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at)")
	if err != nil {
		return fmt.Errorf("failed to create email_verifications index: %v", err)
	}

	log.Printf("Creating indexes for trips")
	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_trips_user_id ON trips(user_id)")
	if err != nil {
		return fmt.Errorf("failed to create trips user_id index: %v", err)
	}

	// Create location index
	log.Printf("Creating index on trips.location")
	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_trips_location ON trips(location)")
	if err != nil {
		return fmt.Errorf("failed to create trips location index: %v", err)
	}

	_, err = TestDB.Exec(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_trips_date_range ON trips(start_date, end_date)")
	if err != nil {
		return fmt.Errorf("failed to create trips date_range index: %v", err)
	}

	log.Printf("All indexes created successfully")
	return nil
}

// CleanTestTables truncates all test tables for a clean test environment
func CleanTestTables(ctx context.Context) error {
	log.Printf("Cleaning test tables")

	// Disable foreign key constraints temporarily
	_, err := TestDB.Exec(ctx, `SET session_replication_role = 'replica';`)
	if err != nil {
		return err
	}

	// Truncate all tables
	_, err = TestDB.Exec(ctx, `
		TRUNCATE TABLE email_verifications, 
		sessions, 
		oauth_accounts, 
		trips, 
		users CASCADE;
	`)

	// Re-enable foreign key constraints
	if err == nil {
		_, err = TestDB.Exec(ctx, `SET session_replication_role = 'origin';`)
	}

	return err
}
