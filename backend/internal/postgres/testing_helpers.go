package postgres

import (
    "database/sql"
    "fmt"
    "testing"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "github.com/stretchr/testify/require"

    "github.com/TommySanDev/gachiakuta-hispano/internal/user"
    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
    "github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

// Test database configuration
const (
    testDBHost     = "localhost"
    testDBPort     = "5432"
    testDBUser     = "postgres"
    testDBPassword = "postgres"
    testDBName     = "gachiakuta_test_db"
)

// SetupTestDB creates a test database connection and ensures tables exist
func SetupTestDB(t *testing.T) *sqlx.DB {
    t.Helper()

    // Connect to postgres database to create test database
    adminConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
        testDBHost, testDBPort, testDBUser, testDBPassword)
    
    adminDB, err := sql.Open("postgres", adminConnStr)
    require.NoError(t, err)
    defer adminDB.Close()

    // Create test database if it doesn't exist
    _, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
    if err != nil && !isDBExistsError(err) {
        require.NoError(t, err)
    }

    // Connect to test database
    testConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)
    
    db, err := sqlx.Connect("postgres", testConnStr)
    require.NoError(t, err)

    // Create tables if they don't exist
    createTestTables(t, db)

    return db
}

// CleanupTestDB cleans all test data from tables
func CleanupTestDB(t *testing.T, db *sqlx.DB) {
    t.Helper()

    tables := []string{
        "vital_instruments",
        "characters",
        "recovery_codes",
        "totp_secrets", 
        "reset_tokens",
        "magic_links",
        "sessions",
        "users",
    }

    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
        require.NoError(t, err)
    }
}

// TeardownTestDB closes the database connection
func TeardownTestDB(t *testing.T, db *sqlx.DB) {
    t.Helper()
    require.NoError(t, db.Close())
}

// CreateTestUser creates a test user with default values
func CreateTestUser(t *testing.T, overrides ...*user.User) *user.User {
    t.Helper()

    // Generate simple, short, unique values
    timestamp := time.Now().UnixNano()
    shortID := timestamp % 100000 // Only last 5 digits for brevity
    
    testUser := &user.User{
        Email:             fmt.Sprintf("test%d@example.com", shortID),
        Username:          fmt.Sprintf("user%d", shortID),
        PasswordHash:      "$2a$10$test.hash.here",
        FirstName:         "Test",
        LastName:          "User",
        Role:              user.RoleUser,
        IsActive:          true,
        EmailVerified:     true,
        MagicLinkEnabled:  false,
        TOTPEnabled:       false,
        CreatedAt:         time.Now(),
        UpdatedAt:         time.Now(),
    }

    // Apply overrides
    if len(overrides) > 0 && overrides[0] != nil {
        override := overrides[0]
        if override.Email != "" {
            testUser.Email = override.Email
        }
        if override.Username != "" {
            testUser.Username = override.Username
        }
        if override.PasswordHash != "" {
            testUser.PasswordHash = override.PasswordHash
        }
        if override.FirstName != "" {
            testUser.FirstName = override.FirstName
        }
        if override.LastName != "" {
            testUser.LastName = override.LastName
        }
        if override.Role != "" {
            testUser.Role = override.Role
        }
        testUser.IsActive = override.IsActive
        testUser.EmailVerified = override.EmailVerified
        testUser.MagicLinkEnabled = override.MagicLinkEnabled
        testUser.TOTPEnabled = override.TOTPEnabled
    }

    return testUser
}

// CreateTestSession creates a test session with default values
func CreateTestSession(t *testing.T, userID uint, overrides ...*user.Session) *user.Session {
    t.Helper()

    testSession := &user.Session{
        ID:        "test-session-id",
        UserID:    userID,
        Token:     "test-token-hash",
        UserAgent: "Test User Agent",
        IPAddress: "127.0.0.1",
        ExpiresAt: time.Now().Add(24 * time.Hour),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Apply overrides
    if len(overrides) > 0 && overrides[0] != nil {
        override := overrides[0]
        if override.ID != "" {
            testSession.ID = override.ID
        }
        if override.Token != "" {
            testSession.Token = override.Token
        }
        if override.UserAgent != "" {
            testSession.UserAgent = override.UserAgent
        }
        if override.IPAddress != "" {
            testSession.IPAddress = override.IPAddress
        }
        if !override.ExpiresAt.IsZero() {
            testSession.ExpiresAt = override.ExpiresAt
        }
    }

    return testSession
}

// CreateTestMagicLink creates a test magic link with default values
func CreateTestMagicLink(t *testing.T, userID uint, overrides ...*user.MagicLink) *user.MagicLink {
    t.Helper()

    testLink := &user.MagicLink{
        UserID:    userID,
        Token:     "test-magic-token",
        Used:      false,
        ExpiresAt: time.Now().Add(15 * time.Minute),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Apply overrides
    if len(overrides) > 0 && overrides[0] != nil {
        override := overrides[0]
        if override.Token != "" {
            testLink.Token = override.Token
        }
        testLink.Used = override.Used
        if !override.ExpiresAt.IsZero() {
            testLink.ExpiresAt = override.ExpiresAt
        }
    }

    return testLink
}

// CreateTestCharacter creates a test character with default values
func CreateTestCharacter(t *testing.T, overrides ...*character.Character) *character.Character {
    t.Helper()

    testCharacter := &character.Character{
        Name:            "Test Character",
        NameJapanese:    "テストキャラクター",
        MainImage:       "/images/test-character.jpg",
        Description:     "A test character for unit testing",
        Species:         "Human",
        Gender:          "Unknown",
        Age:             20,
        Height:          "170cm",
        Status:          "Alive",
        Affiliation:     "Test Affiliation",
        Occupation:      "Test Occupation",
        BirthDate:       "Unknown",
        BirthPlace:      "Test Place",
        Relatives:       "None",
        FirstAppearance: 1,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // Apply overrides
    if len(overrides) > 0 && overrides[0] != nil {
        override := overrides[0]
        if override.Name != "" {
            testCharacter.Name = override.Name
        }
        if override.NameJapanese != "" {
            testCharacter.NameJapanese = override.NameJapanese
        }
        if override.MainImage != "" {
            testCharacter.MainImage = override.MainImage
        }
        if override.Description != "" {
            testCharacter.Description = override.Description
        }
        if override.Species != "" {
            testCharacter.Species = override.Species
        }
        if override.Gender != "" {
            testCharacter.Gender = override.Gender
        }
        if override.Age != 0 {
            testCharacter.Age = override.Age
        }
        if override.Height != "" {
            testCharacter.Height = override.Height
        }
        if override.Status != "" {
            testCharacter.Status = override.Status
        }
        if override.Affiliation != "" {
            testCharacter.Affiliation = override.Affiliation
        }
        if override.Occupation != "" {
            testCharacter.Occupation = override.Occupation
        }
        if override.BirthDate != "" {
            testCharacter.BirthDate = override.BirthDate
        }
        if override.BirthPlace != "" {
            testCharacter.BirthPlace = override.BirthPlace
        }
        if override.Relatives != "" {
            testCharacter.Relatives = override.Relatives
        }
        if override.FirstAppearance != 0 {
            testCharacter.FirstAppearance = override.FirstAppearance
        }
    }

    return testCharacter
}

// CreateTestVitalInstrument creates a test vital instrument with default values
func CreateTestVitalInstrument(t *testing.T, characterID *uint, overrides ...*vitalinstrument.VitalInstrument) *vitalinstrument.VitalInstrument {
    t.Helper()

    testInstrument := &vitalinstrument.VitalInstrument{
        Name:            "Test Instrument",
        MainImage:       "/images/test-instrument.jpg",
        Description:     "A test vital instrument for unit testing",
        Powers:          "Test powers and abilities",
        CharacterID:     characterID,
        FirstAppearance: 1,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // Apply overrides
    if len(overrides) > 0 && overrides[0] != nil {
        override := overrides[0]
        if override.Name != "" {
            testInstrument.Name = override.Name
        }
        if override.MainImage != "" {
            testInstrument.MainImage = override.MainImage
        }
        if override.Description != "" {
            testInstrument.Description = override.Description
        }
        if override.Powers != "" {
            testInstrument.Powers = override.Powers
        }
        if override.CharacterID != nil {
            testInstrument.CharacterID = override.CharacterID
        }
        if override.FirstAppearance != 0 {
            testInstrument.FirstAppearance = override.FirstAppearance
        }
    }

    return testInstrument
}

// Helper functions

func isDBExistsError(err error) bool {
    return err != nil && 
           (err.Error() == fmt.Sprintf(`pq: database "%s" already exists`, testDBName))
}

func createTestTables(t *testing.T, db *sqlx.DB) {
    t.Helper()

    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        username VARCHAR(100) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        first_name VARCHAR(100) NOT NULL,
        last_name VARCHAR(100) NOT NULL,
        role VARCHAR(50) NOT NULL DEFAULT 'user',
        is_active BOOLEAN NOT NULL DEFAULT true,
        email_verified BOOLEAN NOT NULL DEFAULT false,
        magic_link_enabled BOOLEAN NOT NULL DEFAULT false,
        totp_enabled BOOLEAN NOT NULL DEFAULT false,
        last_login_at TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        deleted_at TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS sessions (
        id VARCHAR(255) PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        token VARCHAR(255) NOT NULL,
        user_agent TEXT,
        ip_address INET,
        expires_at TIMESTAMP NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS magic_links (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        token VARCHAR(255) NOT NULL,
        used BOOLEAN NOT NULL DEFAULT false,
        expires_at TIMESTAMP NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS reset_tokens (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        token VARCHAR(255) NOT NULL,
        used BOOLEAN NOT NULL DEFAULT false,
        expires_at TIMESTAMP NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS totp_secrets (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        secret VARCHAR(255) NOT NULL,
        verified BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        UNIQUE(user_id)
    );

    CREATE TABLE IF NOT EXISTS recovery_codes (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        code VARCHAR(255) NOT NULL,
        used BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS characters (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        name_japanese VARCHAR(255),
        main_image TEXT NOT NULL,
        description TEXT NOT NULL,
        species VARCHAR(100),
        gender VARCHAR(50),
        age INTEGER,
        height VARCHAR(50),
        status VARCHAR(100) NOT NULL,
        affiliation VARCHAR(255),
        occupation VARCHAR(255),
        birth_date VARCHAR(100),
        birth_place VARCHAR(255),
        relatives TEXT,
        first_appearance INTEGER NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        deleted_at TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS vital_instruments (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        main_image TEXT NOT NULL,
        description TEXT NOT NULL,
        powers TEXT,
        character_id INTEGER REFERENCES characters(id),
        first_appearance INTEGER NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        deleted_at TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
    CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
    CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
    CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
    CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
    CREATE INDEX IF NOT EXISTS idx_magic_links_user_id ON magic_links(user_id);
    CREATE INDEX IF NOT EXISTS idx_magic_links_token ON magic_links(token);
    CREATE INDEX IF NOT EXISTS idx_reset_tokens_user_id ON reset_tokens(user_id);
    CREATE INDEX IF NOT EXISTS idx_reset_tokens_token ON reset_tokens(token);
    CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name);
    CREATE INDEX IF NOT EXISTS idx_characters_status ON characters(status);
    CREATE INDEX IF NOT EXISTS idx_characters_affiliation ON characters(affiliation);
    CREATE INDEX IF NOT EXISTS idx_vital_instruments_name ON vital_instruments(name);
    CREATE INDEX IF NOT EXISTS idx_vital_instruments_character_id ON vital_instruments(character_id);
    `

    _, err := db.Exec(schema)
    require.NoError(t, err)
}
