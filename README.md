# Gachiakuta Hispano

Una fan page hispana completa para el manga Gachiakuta con backend en Go y frontend en Astro.
Incluye autenticación completa, gestión de contenido, comentarios, favoritos y sistema de administración.

---

## 🚀 Stack Tecnológico

### Backend

* **Go 1.19.8** – API REST robusta
* **PostgreSQL** – Base de datos principal
* **Chi Router** – Routing rápido y minimalista
* **PASETO** – Tokens seguros para autenticación
* **SQLx** – ORM ligero para PostgreSQL
* **Zap** – Logging estructurado
* **TOTP/2FA** – Autenticación de dos factores
* **Magic Links** – Autenticación sin contraseña

### Frontend

* **Astro 5.5.4** – Framework moderno con Server Islands
* **TailwindCSS 4.1** – Diseño responsivo y moderno
* **TypeScript** – Tipado estático
* **Preact Signals** – Estado reactivo
* **Node.js** – Adapter para funciones server-side

### Desarrollo & DevOps

* **Docker** – PostgreSQL y MailHog containerizados
* **MailHog** – Testing de emails en desarrollo
* **Sharp** – Optimización de imágenes

---

## 📋 Requisitos Previos

* Go 1.19+
* Node.js 18+ + `pnpm`
* Docker & Docker Compose
* PostgreSQL (vía Docker o instalación local)

---

## ⚡ Inicio Rápido

### 1. Clonar y Configurar

```bash
# Clonar repositorio
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
```

### 2. Configurar Variables de Entorno

```bash
cd backend
cp .env.example .env
```

Editar `backend/.env`:

```env
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gachiakuta_fan_db
DB_HOST=localhost
DB_PORT=5432
DB_SSLMODE=disable

# IMPORTANTE: Generar clave PASETO segura
PASETO_SECRET=12345678901234567890123456789012

FRONTEND_URL=http://localhost:4321
BACKEND_URL=http://localhost:8080

SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_FROM=noreply@gachiakuta-hispano.com
```

⚠️ En producción, genera una clave PASETO segura con:

```bash
head -c 32 /dev/urandom | base64
```

---

### 3. Levantar Servicios con Docker

#### PostgreSQL

```bash
docker run --name gachiakuta-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gachiakuta_fan_db \
  -p 5432:5432 \
  -d postgres:latest
```

#### MailHog (para testing de emails)

```bash
docker run --name gachiakuta-mailhog \
  -p 1025:1025 \
  -p 8025:8025 \
  -d mailhog/mailhog
```

Verificar que estén corriendo:

```bash
docker ps
```

Servicios disponibles:

* PostgreSQL: `localhost:5432`
* MailHog Web UI: [http://localhost:8025](http://localhost:8025)
* MailHog SMTP: `localhost:1025`

---

### 4. Configurar Base de Datos

```bash
psql -h localhost -U postgres -p 5432
```

En psql:

```sql
CREATE DATABASE gachiakuta_fan_db;
\q
```

Ejecutar schema desde el directorio raíz:

```bash
psql -h localhost -U postgres -d gachiakuta_fan_db -f backend/sql/schema.sql
```

Usuario admin creado automáticamente:

* Email: `admin@gachiakuta-hispano.com`
* Password: `admin123`
* Role: `admin`

---

### 5. Generar Mapeo de Imágenes (Frontend)

```bash
cd frontend/gachiakuta-astro
pnpm run generate-image-map
```

Esto genera/actualiza `src/lib/image-map.ts`, mapeando imágenes de:

* `src/assets/chars/`
* `src/assets/vi/`
* `src/assets/chapters/`

---

### 6. Ejecutar el Proyecto

**Terminal 1 - Backend:**

```bash
cd backend
go run main.go
```

Servidor: [http://localhost:8080](http://localhost:8080)

**Terminal 2 - Frontend:**

```bash
cd frontend/gachiakuta-astro
pnpm run dev
```

Aplicación: [http://localhost:4321](http://localhost:4321)

---

## 🎯 Endpoints Principales

### 🔐 Autenticación

```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/magic-link
GET  /api/auth/magic-link
POST /api/auth/reset-password
```

### 👤 Usuarios

```
GET  /api/users/me
PUT  /api/users/me
POST /api/users/me/change-password
```

### 🔒 2FA

```
POST   /api/2fa/setup
POST   /api/2fa/verify
DELETE /api/2fa/disable
POST   /api/2fa/recovery-codes
```

### 📚 Contenido Público

```
GET /api/characters
GET /api/characters/{id}
GET /api/chapters
GET /api/chapters/{id}
GET /api/vital-instruments
```

### 💬 Comentarios

```
GET    /api/comments/chapter/{id}
POST   /api/comments
PUT    /api/comments/{id}
DELETE /api/comments/{id}
```

### ⭐ Favoritos

```
GET    /api/favorites
POST   /api/favorites
DELETE /api/favorites
```

### 👑 Admin

```
GET    /api/admin/users
POST   /api/admin/users
PUT    /api/admin/users/{id}
DELETE /api/admin/users/{id}
POST   /api/admin/users/{id}/reset-password
DELETE /api/admin/users/{id}/disable-2fa
```

---

## 🏗️ Arquitectura del Proyecto

```text
gachiakuta-hispano/
├── backend/
│   ├── cmd/
│   ├── config/
│   ├── internal/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── user/
│   │   ├── character/
│   │   ├── chapter/
│   │   ├── comment/
│   │   └── favorite/
│   ├── sql/
│   ├── .env.example
│   └── main.go
│
├── frontend/gachiakuta-astro/
│   ├── src/
│   │   ├── components/
│   │   │   ├── auth/
│   │   │   ├── admin/
│   │   │   ├── profile/
│   │   │   └── comments/
│   │   ├── layouts/
│   │   ├── pages/
│   │   │   ├── auth/
│   │   │   ├── admin/
│   │   │   ├── characters/
│   │   │   └── chapters/
│   │   ├── lib/
│   │   │   ├── auth-store.ts
│   │   │   ├── auth-server.ts
│   │   │   └── api-client.ts
│   │   ├── types/
│   │   └── assets/
│   └── package.json
│
└── docker-compose.yml
```

---

## 📧 Testing de Emails

MailHog:

* Web UI: [http://localhost:8025](http://localhost:8025)
* Funcionalidades:

  * Ver todos los emails enviados
  * Testing de magic links
  * Testing de verificación de email
  * Testing de reset de contraseñas

---

## 🔒 Seguridad

* PASETO v2 para tokens seguros
* bcrypt para hash de contraseñas
* TOTP/2FA con Google Authenticator
* CORS configurado
* Rate limiting en autenticación
* Validación de inputs en frontend y backend
* Protección contra SQL injection con SQLx

---

## 🚦 Estructura de Roles

### 👑 Admin

* Acceso completo al sistema
* Gestión de usuarios (CRUD + roles)
* Reset de contraseñas
* Desactivación de 2FA

### 👤 User

* Lectura de contenido público
* Gestión de perfil
* Configuración de 2FA / Magic Links

---

## 📄 Licencia

Este proyecto es una fan page no comercial del manga **Gachiakuta**.
Todos los derechos del contenido original pertenecen a sus respectivos autores.

---

¡Disfruta explorando el mundo de Gachiakuta! 🌼⚔️

