Gachiakuta Hispano
Una fan page hispana completa para el manga Gachiakuta con backend en Go y frontend en Astro. Incluye autenticación completa, gestión de contenido, comentarios, favoritos y sistema de administración.
🚀 Stack Tecnológico
Backend

Go 1.19.8 - API REST robusta
PostgreSQL - Base de datos principal
Chi Router - Routing rápido y minimalista
PASETO - Tokens seguros para autenticación
SQLx - ORM ligero para PostgreSQL
Zap - Logging estructurado
TOTP/2FA - Autenticación de dos factores
Magic Links - Autenticación sin contraseña

Frontend

Astro 5.5.4 - Framework moderno con Server Islands
TailwindCSS 4.1 - Diseño responsivo y moderno
TypeScript - Tipado estático
Preact Signals - Estado reactivo
Node.js - Adapter para funciones server-side

Desarrollo & DevOps

Docker - PostgreSQL y MailHog containerizados
MailHog - Testing de emails en desarrollo
Sharp - Optimización de imágenes

📋 Requisitos Previos

Go 1.19+
Node.js 18+ + pnpm
Docker & Docker Compose
PostgreSQL (vía Docker o instalación local)

⚡ Inicio Rápido
1. Clonar y Configurar
bash# Clonar repositorio
git clone <repo-url>
cd gachiakuta-hispano

# Instalar dependencias del frontend
cd frontend/gachiakuta-astro
pnpm install
cd ../../

# Instalar dependencias del backend
cd backend
go mod download
cd ../
2. Configurar Variables de Entorno
bash# Backend - Copiar y configurar .env
cd backend
cp .env.example .env
Editar backend/.env:
bash# Database Configuration
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gachiakuta_fan_db
DB_HOST=localhost
DB_PORT=5432
DB_SSLMODE=disable

# IMPORTANTE: Generar clave PASETO segura
# Ejecutar: head -c 32 /dev/urandom | base64 o utilizar placeholder NO Seguro fuera de pruebas>
PASETO_SECRET=12345678901234567890123456789012


# URLs del proyecto
FRONTEND_URL=http://localhost:4321
BACKEND_URL=http://localhost:8080

# SMTP para desarrollo (MailHog)
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_FROM=noreply@gachiakuta-hispano.com

# El resto de valores pueden quedarse como están
⚠️ IMPORTANTE: En producción, generar una clave PASETO segura:
bash# head -c 32 /dev/urandom | base64

3. Levantar Servicios con Docker (Manualmente)
PostgreSQL:
bash# Opción 1: PostgreSQL con Docker
docker run --name gachiakuta-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gachiakuta_fan_db \
  -p 5432:5432 \
  -d postgres:latest

# Opción 2: PostgreSQL local (si ya tienes instalado)
# Crear la base de datos manualmente con tu instalación local

MailHog (para testing de emails):
bash# Levantar MailHog
docker run --name gachiakuta-mailhog \
  -p 1025:1025 \
  -p 8025:8025 \
  -d mailhog/mailhog

# Verificar que estén corriendo
docker ps
Servicios disponibles:

PostgreSQL: localhost:5432
MailHog Web UI: http://localhost:8025
MailHog SMTP: localhost:1025

4. Configurar Base de Datos
bash# Conectar a PostgreSQL
psql -h localhost -U postgres -p 5432

# Crear la base de datos
CREATE DATABASE gachiakuta_fan_db;
\q

# Ejecutar schema (desde directorio raíz del proyecto)
psql -h localhost -U postgres -d gachiakuta_fan_db -f backend/sql/schema.sql
Usuario admin creado automáticamente:

Email: admin@gachiakuta-hispano.com
Password: admin123
Role: admin

5. Generar Mapeo de Imágenes (Frontend)
bash# cd frontend/gachiakuta-astro

# Generar mapeo automático de imágenes
pnpm run generate-image-map

# Esto crea/actualiza: src/lib/image-map.ts
Este script escanea las carpetas src/assets/chars/, src/assets/vi/ y src/assets/chapters/ y genera automáticamente el mapeo de imágenes necesario para Astro.

6. Ejecutar el Proyecto
Terminal 1 - Backend:
bash# cd backend
go run main.go

# Servidor corriendo en: http://localhost:8080
Terminal 2 - Frontend:
bash# cd frontend/gachiakuta-astro
pnpm run dev

# Aplicación corriendo en: http://localhost:4321
🎯 Endpoints Principales
🔐 Autenticación
POST /api/auth/register       # Registro de usuarios
POST /api/auth/login          # Login con email/password
POST /api/auth/logout         # Cerrar sesión
POST /api/auth/magic-link     # Solicitar magic link
GET  /api/auth/magic-link     # Login con magic link
POST /api/auth/reset-password # Solicitar reset de password
👤 Usuarios
GET  /api/users/me            # Perfil del usuario actual
PUT  /api/users/me            # Actualizar perfil
POST /api/users/me/change-password # Cambiar contraseña
🔒 2FA
POST /api/2fa/setup          # Configurar 2FA
POST /api/2fa/verify         # Verificar y activar 2FA
DELETE /api/2fa/disable      # Desactivar 2FA
POST /api/2fa/recovery-codes # Generar códigos de recuperación
📚 Contenido Público
GET /api/characters          # Lista de personajes
GET /api/characters/{id}     # Detalle de personaje
GET /api/chapters            # Lista de capítulos  
GET /api/chapters/{id}       # Detalle de capítulo
GET /api/vital-instruments   # Lista de instrumentos vitales
💬 Comentarios
GET  /api/comments/chapter/{id}  # Comentarios de un capítulo
POST /api/comments               # Crear comentario (autenticado)
PUT  /api/comments/{id}          # Editar comentario (autor/admin)
DELETE /api/comments/{id}        # Eliminar comentario (autor/admin)
⭐ Favoritos
GET    /api/favorites            # Favoritos del usuario
POST   /api/favorites            # Agregar favorito
DELETE /api/favorites            # Eliminar favorito
👑 Admin
GET    /api/admin/users          # Lista de usuarios
POST   /api/admin/users          # Crear usuario
PUT    /api/admin/users/{id}     # Actualizar usuario
DELETE /api/admin/users/{id}     # Eliminar usuario
POST   /api/admin/users/{id}/reset-password  # Reset password (admin)
DELETE /api/admin/users/{id}/disable-2fa     # Desactivar 2FA (admin)
🏗️ Arquitectura del Proyecto
gachiakuta-hispano/
├── backend/                     # API Go
│   ├── cmd/                     # Comandos y scripts
│   ├── config/                  # Configuración (DB, Auth, SMTP)
│   ├── internal/                # Lógica de negocio
│   │   ├── handler/             # HTTP handlers
│   │   ├── middleware/          # Auth, logging, CORS
│   │   ├── user/                # Módulo de usuarios y auth
│   │   ├── character/           # Módulo de personajes
│   │   ├── chapter/             # Módulo de capítulos
│   │   ├── comment/             # Módulo de comentarios
│   │   └── favorite/            # Módulo de favoritos
│   ├── sql/                     # Schemas y migraciones
│   ├── .env.example             # Variables de entorno
│   └── main.go                  # Punto de entrada
│
├── frontend/gachiakuta-astro/   # Frontend Astro
│   ├── src/
│   │   ├── components/          # Componentes reutilizables
│   │   │   ├── auth/            # Componentes de autenticación
│   │   │   ├── admin/           # Panel de administración
│   │   │   ├── profile/         # Gestión de perfil
│   │   │   └── comments/        # Sistema de comentarios
│   │   ├── layouts/             # Layouts (TV, Header, Admin)
│   │   ├── pages/               # Páginas de la aplicación
│   │   │   ├── auth/            # Páginas de autenticación
│   │   │   ├── admin/           # Panel administrativo
│   │   │   ├── characters/      # Páginas de personajes
│   │   │   └── chapters/        # Páginas de capítulos
│   │   ├── lib/                 # Utilidades y APIs
│   │   │   ├── auth-store.ts    # Store de autenticación
│   │   │   ├── auth-server.ts   # Validación server-side
│   │   │   └── api-client.ts    # Cliente HTTP
│   │   ├── types/               # Definiciones TypeScript
│   │   └── assets/              # Imágenes y recursos
│   └── package.json
│
└── docker-compose.yml           # PostgreSQL + MailHog

Funciones Implementadas -
📧 Testing de Emails

MailHog está configurado para testing de emails en desarrollo:

Interfaz Web: http://localhost:8025
Funcionalidades:

Ver todos los emails enviados
Testing de magic links
Testing de verificación de email
Testing de reset de contraseñas


🔒 Seguridad

PASETO v2 para tokens seguros
bcrypt para hash de contraseñas
TOTP/2FA con Google Authenticator
CORS configurado
Rate limiting en autenticación
Validación de inputs en frontend y backend
SQL injection protection con SQLx

🚦 Estructura de Roles
👑 Admin

Acceso completo al sistema
Gestión de usuarios (CRUD + roles)
Reset de contraseñas de usuarios
Desactivación de 2FA de usuarios

👤 User

Lectura de contenido público
Gestión de perfil propio
Configuración de 2FA/Magic Links

📄 Licencia
Este proyecto es una fan page no comercial del manga Gachiakuta. Todos los derechos del contenido original pertenecen a sus respectivos autores.

¡Disfruta explorando el mundo de Gachiakuta! 🎌⚔️
