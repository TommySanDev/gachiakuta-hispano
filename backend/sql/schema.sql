-- =================================================================
-- GACHIAKUTA HISPANO - COMPLETE DATABASE SCHEMA
-- =================================================================
-- Este script es idempotente y puede ejecutarse múltiples veces
-- Genera todas las tablas necesarias para el proyecto

-- Habilitar extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =================================================================
-- USERS Y AUTHENTICATION TABLES
-- =================================================================

-- Tabla principal de usuarios
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'editor', 'user')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    magic_link_enabled BOOLEAN NOT NULL DEFAULT false,
    totp_enabled BOOLEAN NOT NULL DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Magic links para autenticación sin contraseña
CREATE TABLE IF NOT EXISTS magic_links (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(64) UNIQUE NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tokens para reset de contraseña
CREATE TABLE IF NOT EXISTS reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(64) UNIQUE NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Secretos TOTP para 2FA
CREATE TABLE IF NOT EXISTS totp_secrets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(32) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Códigos de recuperación para 2FA
CREATE TABLE IF NOT EXISTS recovery_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tokens de verificación de email
CREATE TABLE IF NOT EXISTS email_verifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(64) UNIQUE NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =================================================================
-- CONTENT TABLES (Characters, Chapters, Vital Instruments)
-- =================================================================

-- Tabla de personajes
CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    name_japanese VARCHAR(255),
    main_image VARCHAR(500),
    description TEXT,
    species VARCHAR(100),
    gender VARCHAR(50),
    age INTEGER,
    height VARCHAR(50),
    status VARCHAR(100),
    affiliation VARCHAR(255),
    occupation VARCHAR(255),
    birth_date VARCHAR(100),
    birth_place VARCHAR(255),
    relatives TEXT,
    first_appearance INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Tabla de capítulos
CREATE TABLE IF NOT EXISTS chapters (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    number INTEGER NOT NULL UNIQUE,
    image VARCHAR(500),
    synopsis TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Tabla de instrumentos vitales
CREATE TABLE IF NOT EXISTS vital_instruments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    main_image VARCHAR(500),
    description TEXT,
    powers TEXT,
    character_id INTEGER REFERENCES characters(id) ON DELETE SET NULL,
    first_appearance INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =================================================================
-- USER INTERACTION TABLES (Comments, Favorites)
-- =================================================================

-- Tabla de comentarios en capítulos
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chapter_id INTEGER NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Tabla de favoritos de usuarios
CREATE TABLE IF NOT EXISTS favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('character', 'chapter', 'vital_instrument')),
    entity_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint para evitar duplicados
    UNIQUE(user_id, entity_type, entity_id)
);

-- =================================================================
-- INDEXES PARA PERFORMANCE
-- =================================================================

-- Users indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Authentication indexes
CREATE INDEX IF NOT EXISTS idx_magic_links_token ON magic_links(token);
CREATE INDEX IF NOT EXISTS idx_magic_links_expires_at ON magic_links(expires_at);
CREATE INDEX IF NOT EXISTS idx_magic_links_user_id ON magic_links(user_id);

CREATE INDEX IF NOT EXISTS idx_reset_tokens_token ON reset_tokens(token);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_expires_at ON reset_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_user_id ON reset_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at);
CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);

CREATE INDEX IF NOT EXISTS idx_totp_secrets_user_id ON totp_secrets(user_id);
CREATE INDEX IF NOT EXISTS idx_recovery_codes_user_id ON recovery_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_recovery_codes_code ON recovery_codes(code);

-- Content indexes
CREATE INDEX IF NOT EXISTS idx_characters_deleted_at ON characters(deleted_at);
CREATE INDEX IF NOT EXISTS idx_characters_species ON characters(species);
CREATE INDEX IF NOT EXISTS idx_characters_status ON characters(status);
CREATE INDEX IF NOT EXISTS idx_characters_affiliation ON characters(affiliation);

CREATE INDEX IF NOT EXISTS idx_chapters_deleted_at ON chapters(deleted_at);
CREATE INDEX IF NOT EXISTS idx_chapters_number ON chapters(number);

CREATE INDEX IF NOT EXISTS idx_vital_instruments_deleted_at ON vital_instruments(deleted_at);
CREATE INDEX IF NOT EXISTS idx_vital_instruments_character_id ON vital_instruments(character_id);

-- User interaction indexes
CREATE INDEX IF NOT EXISTS idx_comments_chapter_id ON comments(chapter_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments(deleted_at);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_entity_type ON favorites(entity_type);
CREATE INDEX IF NOT EXISTS idx_favorites_entity_id ON favorites(entity_id);

-- =================================================================
-- FUNCTIONS Y TRIGGERS PARA UPDATED_AT
-- =================================================================

-- Función para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para updated_at
DO $$
BEGIN
    -- Users
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_users_updated_at') THEN
        CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Magic links
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_magic_links_updated_at') THEN
        CREATE TRIGGER update_magic_links_updated_at BEFORE UPDATE ON magic_links
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Reset tokens
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_reset_tokens_updated_at') THEN
        CREATE TRIGGER update_reset_tokens_updated_at BEFORE UPDATE ON reset_tokens
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- TOTP secrets
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_totp_secrets_updated_at') THEN
        CREATE TRIGGER update_totp_secrets_updated_at BEFORE UPDATE ON totp_secrets
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Recovery codes
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_recovery_codes_updated_at') THEN
        CREATE TRIGGER update_recovery_codes_updated_at BEFORE UPDATE ON recovery_codes
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Email verifications
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_email_verifications_updated_at') THEN
        CREATE TRIGGER update_email_verifications_updated_at BEFORE UPDATE ON email_verifications
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Characters
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_characters_updated_at') THEN
        CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Chapters
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_chapters_updated_at') THEN
        CREATE TRIGGER update_chapters_updated_at BEFORE UPDATE ON chapters
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Vital instruments
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_vital_instruments_updated_at') THEN
        CREATE TRIGGER update_vital_instruments_updated_at BEFORE UPDATE ON vital_instruments
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- Comments
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_comments_updated_at') THEN
        CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END
$$;

-- =================================================================
-- INSERTAR USUARIO ADMIN POR DEFECTO
-- =================================================================

-- Insertar admin user (password: "admin123" - cambiar en producción)
-- Password hash para "admin123" usando bcrypt cost 10
INSERT INTO users (
    email, 
    username, 
    password_hash, 
    first_name, 
    last_name, 
    role, 
    is_active, 
    email_verified,
    magic_link_enabled
) VALUES (
    'admin@gachiakuta-hispano.com',
    'admin',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    'Admin',
    'Gachiakuta',
    'admin',
    true,
    true,
    true
) ON CONFLICT (email) DO NOTHING;

-- =================================================================
-- VERIFICAR ESQUEMA COMPLETADO
-- =================================================================

DO $$
DECLARE
    table_count INTEGER;
    index_count INTEGER;
    trigger_count INTEGER;
BEGIN
    -- Contar tablas creadas
    SELECT COUNT(*) INTO table_count 
    FROM information_schema.tables 
    WHERE table_schema = 'public' 
    AND table_name IN (
        'users', 'magic_links', 'reset_tokens', 'totp_secrets', 
        'recovery_codes', 'email_verifications', 'characters', 
        'chapters', 'vital_instruments', 'comments', 'favorites'
    );
    
    -- Contar índices creados
    SELECT COUNT(*) INTO index_count 
    FROM pg_indexes 
    WHERE schemaname = 'public' 
    AND indexname LIKE 'idx_%';
    
    -- Contar triggers creados
    SELECT COUNT(*) INTO trigger_count 
    FROM pg_trigger 
    WHERE tgname LIKE 'update_%_updated_at';
    
    RAISE NOTICE '=================================================================';
    RAISE NOTICE 'GACHIAKUTA HISPANO - SCHEMA INSTALLATION COMPLETED';
    RAISE NOTICE '=================================================================';
    RAISE NOTICE 'Tables created: % / 11', table_count;
    RAISE NOTICE 'Indexes created: %', index_count;
    RAISE NOTICE 'Triggers created: % / 9', trigger_count;
    RAISE NOTICE 'Admin user: admin@gachiakuta-hispano.com (password: admin123)';
    RAISE NOTICE '=================================================================';
    
    IF table_count = 11 THEN
        RAISE NOTICE '✅ All tables created successfully!';
    ELSE
        RAISE WARNING '⚠️  Some tables may be missing. Expected 11, found %', table_count;
    END IF;
END
$$;
