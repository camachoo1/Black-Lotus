package repositories_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"black-lotus/db"
	"black-lotus/internal/auth/models"
	"black-lotus/internal/auth/repositories"
)

func TestCreateOAuthAccount(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create user repository to create a test user first
	userRepo := repositories.NewUserRepository(db.TestDB)

	// Create a test user
	userInput := models.CreateUserInput{
		Name:     "OAuth Test User",
		Email:    "oauthtest@example.com",
		Password: stringPtr("Password123!"),
	}

	hashedPassword := "hashed_password"

	user, err := userRepo.CreateUser(context.Background(), userInput, &hashedPassword)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create OAuth repository
	repo := repositories.NewOAuthRepository(db.TestDB)

	t.Run("Create New OAuth Account", func(t *testing.T) {
		// Create test account data
		expiresAt := time.Now().Add(time.Hour * 24)
		account := models.OAuthAccount{
			ProviderID:     "google",
			ProviderUserID: "12345",
			UserID:         user.ID,
			AccessToken:    "access_token_123",
			RefreshToken:   "refresh_token_123",
			ExpiresAt:      expiresAt,
		}

		// Create the account
		err := repo.CreateOAuthAccount(context.Background(), account)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify it was created by fetching it
		fetchedAccount, err := repo.GetOAuthAccount(context.Background(), "google", "12345")
		if err != nil {
			t.Fatalf("Failed to fetch created account: %v", err)
		}

		if fetchedAccount == nil {
			t.Fatal("Expected account to be returned, got nil")
		}

		if fetchedAccount.ProviderID != account.ProviderID {
			t.Errorf("Expected provider ID %s, got %s", account.ProviderID, fetchedAccount.ProviderID)
		}

		if fetchedAccount.ProviderUserID != account.ProviderUserID {
			t.Errorf("Expected provider user ID %s, got %s", account.ProviderUserID, fetchedAccount.ProviderUserID)
		}

		if fetchedAccount.UserID != account.UserID {
			t.Errorf("Expected user ID %s, got %s", account.UserID, fetchedAccount.UserID)
		}

		if fetchedAccount.AccessToken != account.AccessToken {
			t.Errorf("Expected access token %s, got %s", account.AccessToken, fetchedAccount.AccessToken)
		}

		if fetchedAccount.RefreshToken != account.RefreshToken {
			t.Errorf("Expected refresh token %s, got %s", account.RefreshToken, fetchedAccount.RefreshToken)
		}

		if fetchedAccount.ExpiresAt.Unix() != account.ExpiresAt.Unix() {
			t.Errorf("Expected expires at %v, got %v", account.ExpiresAt, fetchedAccount.ExpiresAt)
		}

		if fetchedAccount.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}

		if fetchedAccount.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}
	})

	t.Run("Update Existing OAuth Account", func(t *testing.T) {
		// Create initial account data
		expiresAt := time.Now().Add(time.Hour * 24)
		initialAccount := models.OAuthAccount{
			ProviderID:     "github",
			ProviderUserID: "67890",
			UserID:         user.ID,
			AccessToken:    "initial_access_token",
			RefreshToken:   "initial_refresh_token",
			ExpiresAt:      expiresAt,
		}

		// Create the initial account
		err := repo.CreateOAuthAccount(context.Background(), initialAccount)
		if err != nil {
			t.Fatalf("Failed to create initial account: %v", err)
		}

		// Create updated account data
		newExpiresAt := time.Now().Add(time.Hour * 48)
		updatedAccount := models.OAuthAccount{
			ProviderID:     "github",
			ProviderUserID: "67890",
			UserID:         user.ID,
			AccessToken:    "updated_access_token",
			RefreshToken:   "updated_refresh_token",
			ExpiresAt:      newExpiresAt,
		}

		// Update the account
		err = repo.CreateOAuthAccount(context.Background(), updatedAccount)
		if err != nil {
			t.Fatalf("Failed to update account: %v", err)
		}

		// Verify it was updated by fetching it
		fetchedAccount, err := repo.GetOAuthAccount(context.Background(), "github", "67890")
		if err != nil {
			t.Fatalf("Failed to fetch updated account: %v", err)
		}

		if fetchedAccount.AccessToken != updatedAccount.AccessToken {
			t.Errorf("Expected access token %s, got %s", updatedAccount.AccessToken, fetchedAccount.AccessToken)
		}

		if fetchedAccount.RefreshToken != updatedAccount.RefreshToken {
			t.Errorf("Expected refresh token %s, got %s", updatedAccount.RefreshToken, fetchedAccount.RefreshToken)
		}

		if fetchedAccount.ExpiresAt.Unix() != updatedAccount.ExpiresAt.Unix() {
			t.Errorf("Expected expires at %v, got %v", updatedAccount.ExpiresAt, fetchedAccount.ExpiresAt)
		}
	})
}

func TestGetOAuthAccount(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create user repository to create a test user first
	userRepo := repositories.NewUserRepository(db.TestDB)

	// Create a test user
	userInput := models.CreateUserInput{
		Name:     "OAuth Get Test User",
		Email:    "oauthgettest@example.com",
		Password: stringPtr("Password123!"),
	}

	hashedPassword := "hashed_password"

	user, err := userRepo.CreateUser(context.Background(), userInput, &hashedPassword)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create OAuth repository
	repo := repositories.NewOAuthRepository(db.TestDB)

	// Create a test OAuth account
	expiresAt := time.Now().Add(time.Hour * 24)
	account := models.OAuthAccount{
		ProviderID:     "facebook",
		ProviderUserID: "fb12345",
		UserID:         user.ID,
		AccessToken:    "fb_access_token",
		RefreshToken:   "fb_refresh_token",
		ExpiresAt:      expiresAt,
	}

	err = repo.CreateOAuthAccount(context.Background(), account)
	if err != nil {
		t.Fatalf("Failed to create test OAuth account: %v", err)
	}

	t.Run("Existing OAuth Account", func(t *testing.T) {
		fetchedAccount, err := repo.GetOAuthAccount(context.Background(), "facebook", "fb12345")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if fetchedAccount == nil {
			t.Fatal("Expected account to be returned, got nil")
		}

		if fetchedAccount.ProviderID != account.ProviderID {
			t.Errorf("Expected provider ID %s, got %s", account.ProviderID, fetchedAccount.ProviderID)
		}

		if fetchedAccount.ProviderUserID != account.ProviderUserID {
			t.Errorf("Expected provider user ID %s, got %s", account.ProviderUserID, fetchedAccount.ProviderUserID)
		}

		if fetchedAccount.UserID != account.UserID {
			t.Errorf("Expected user ID %s, got %s", account.UserID, fetchedAccount.UserID)
		}
	})

	t.Run("Non-existent OAuth Account", func(t *testing.T) {
		fetchedAccount, err := repo.GetOAuthAccount(context.Background(), "nonexistent", "provider")

		if err == nil {
			t.Error("Expected error for non-existent OAuth account, got nil")
		}

		if fetchedAccount != nil {
			t.Errorf("Expected nil account, got: %v", fetchedAccount)
		}

		if !isPgxNoRows(err) {
			t.Errorf("Expected pgx.ErrNoRows, got: %v", err)
		}
	})
}

func TestGetUserOAuthAccounts(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create user repository to create a test user first
	userRepo := repositories.NewUserRepository(db.TestDB)

	// Create a test user
	userInput := models.CreateUserInput{
		Name:     "OAuth User Accounts Test",
		Email:    "oauthusertest@example.com",
		Password: stringPtr("Password123!"),
	}

	hashedPassword := "hashed_password"

	user, err := userRepo.CreateUser(context.Background(), userInput, &hashedPassword)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create OAuth repository
	repo := repositories.NewOAuthRepository(db.TestDB)

	// Create multiple test OAuth accounts for the user
	providers := []string{"google", "github", "facebook"}
	for i, provider := range providers {
		expiresAt := time.Now().Add(time.Hour * 24)
		account := models.OAuthAccount{
			ProviderID:     provider,
			ProviderUserID: provider + "_user_" + strconv.Itoa(i),
			UserID:         user.ID,
			AccessToken:    provider + "_access_token",
			RefreshToken:   provider + "_refresh_token",
			ExpiresAt:      expiresAt,
		}

		err = repo.CreateOAuthAccount(context.Background(), account)
		if err != nil {
			t.Fatalf("Failed to create test OAuth account for %s: %v", provider, err)
		}
	}

	t.Run("Get User's OAuth Accounts", func(t *testing.T) {
		accounts, err := repo.GetUserOAuthAccounts(context.Background(), user.ID)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if accounts == nil {
			t.Fatal("Expected accounts to be returned, got nil")
		}

		if len(accounts) != len(providers) {
			t.Errorf("Expected %d accounts, got %d", len(providers), len(accounts))
		}

		// Check that all providers are present
		providerMap := make(map[string]bool)
		for _, account := range accounts {
			providerMap[account.ProviderID] = true
		}

		for _, provider := range providers {
			if !providerMap[provider] {
				t.Errorf("Expected provider %s to be in returned accounts", provider)
			}
		}
	})

	t.Run("User With No OAuth Accounts", func(t *testing.T) {
		// Create another user with no OAuth accounts
		newUserInput := models.CreateUserInput{
			Name:     "No OAuth User",
			Email:    "nooauth@example.com",
			Password: stringPtr("Password123!"),
		}

		newUser, err := userRepo.CreateUser(context.Background(), newUserInput, &hashedPassword)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		accounts, err := repo.GetUserOAuthAccounts(context.Background(), newUser.ID)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if accounts == nil {
			t.Fatal("Expected empty accounts slice to be returned, got nil")
		}

		if len(accounts) != 0 {
			t.Errorf("Expected 0 accounts, got %d", len(accounts))
		}
	})

	t.Run("Non-existent User", func(t *testing.T) {
		nonExistentID := uuid.New()
		accounts, err := repo.GetUserOAuthAccounts(context.Background(), nonExistentID)

		if err != nil {
			t.Errorf("Expected no error for non-existent user, got: %v", err)
		}

		if accounts == nil {
			t.Fatal("Expected empty accounts slice to be returned, got nil")
		}

		if len(accounts) != 0 {
			t.Errorf("Expected 0 accounts, got %d", len(accounts))
		}
	})
}
