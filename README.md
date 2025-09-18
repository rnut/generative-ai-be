# Workshop BE (Fiber)

Simple Go Fiber backend with Authentication, JWT, SQLite, Swagger, and Profile feature.

## Prerequisites
- Go 1.22+
- (Optional) Air for hot reload

## Run
```bash
go run main.go
```
Then open: http://localhost:3000

Swagger UI: http://localhost:3000/swagger/index.html

## Environment Variables
- `PORT` (default 3000)
- `DB_PATH` (default data/app.db)
- `JWT_SECRET` (required in prod; dev fallback used if missing)
- `APP_ENV` (dev|prod, affects future behaviors)

Copy `.env.example` to `.env` and adjust.

## Hot Reload (Air)
Install once:
```bash
go install github.com/air-verse/air@latest
```
Run:
```bash
air
```

## Endpoints (Summary)
- GET `/healthz` - liveness
- POST `/api/v1/auth/register` - register (email, password)
- POST `/api/v1/auth/login` - login → JWT access token
- GET `/api/v1/auth/me` - current user (Bearer token)
- GET `/api/v1/profile` - profile (Bearer token)
- PUT `/api/v1/profile` - update editable profile fields (Bearer token)

## JWT Usage
After login you get:
```
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_in": 900
}
```
Include header:
```
Authorization: Bearer <jwt>
```

## Profile Update Rules
Editable: first_name, last_name, phone
Read-only: email, membership_level, membership_code, points, joined_at
Phone normalized to digits (10 digits required if provided).

## Swagger Generation
```bash
go install github.com/swaggo/swag/cmd/swag@latest
$(go env GOPATH)/bin/swag init -g main.go -o internal/docs
```

## SQLite Data
Database file stored at `data/app.db`. Auto-migration adds new columns.

## Build
```bash
go build -o workshop-be
```

## Graceful Shutdown
SIGINT/SIGTERM → shutdown server and close DB.

## Disclaimer
Dev fallback JWT secret is insecure; ensure `JWT_SECRET` set in production.

### System Context Diagram
```mermaid
flowchart LR
    subgraph ClientSide[Client / Developer]
        U[User / Client App]
        SW[Swagger UI]
    end

    subgraph Backend[Go Fiber API]
        H[HTTP Handlers]
        MW[Auth Middleware]
        SVC[Auth & Profile Service]
        JWT[JWT Library]
        GORM[GORM]
    end

    DB[(SQLite)]

    U -- REST Calls --> H
    SW -- Try APIs --> H
    H --> MW
    MW --> SVC
    SVC --> GORM --> DB
    SVC --> JWT
    JWT -- Validate --> MW
```
คำอธิบาย: ผู้ใช้และ Swagger UI เรียก API → Handlers → Middleware ตรวจ Token → Service จัดการ Logic → GORM เข้าถึง DB และใช้ JWT สำหรับสร้าง/ตรวจสอบ Token.

### Sequence Diagram: Register
```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant A as Fiber API
    participant S as Service
    participant DB as SQLite
    participant B as Bcrypt

    C->>A: POST /auth/register (email,password)
    A->>S: Validate input
    S->>DB: Query existing email
    DB-->>S: none
    S->>B: Hash password
    B-->>S: hash
    S->>DB: Insert user
    DB-->>S: user (id, timestamps)
    S-->>A: RegisterOutput
    A-->>C: 201 Created
```

### Sequence Diagram: Login
```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant A as Fiber API
    participant S as Service
    participant DB as SQLite
    participant J as JWT

    C->>A: POST /auth/login
    A->>S: Validate credentials
    S->>DB: Load user by email
    DB-->>S: user + hash
    S->>S: Compare hash
    S->>J: Generate token (sub,email,exp)
    J-->>S: JWT string
    S-->>A: LoginOutput
    A-->>C: 200 OK (access_token)
```

### Sequence Diagram: Get Profile
```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant A as Fiber API
    participant M as Middleware
    participant S as Service
    participant DB as SQLite
    participant J as JWT

    C->>A: GET /profile (Bearer token)
    A->>M: AuthRequired
    M->>J: Parse token
    J-->>M: Claims(sub,email)
    M->>S: GetProfile(sub)
    S->>DB: SELECT user
    DB-->>S: user
    S-->>A: ProfileResponse
    A-->>C: 200 OK
```

### Sequence Diagram: Update Profile
```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant A as Fiber API
    participant M as Middleware
    participant S as Service
    participant DB as SQLite

    C->>A: PUT /profile {first_name,phone?}
    A->>M: AuthRequired
    M->>S: UpdateProfile(sub,payload)
    S->>DB: SELECT user
    DB-->>S: user
    S->>S: Validate & normalize
    S->>DB: UPDATE editable fields
    DB-->>S: updated user
    S-->>A: ProfileResponse
    A-->>C: 200 OK
```

### Future Diagram Extensions
- Refresh Token Flow
- Password Reset Activity Diagram
- C4 Container Diagram (เมื่อมี external services)

---