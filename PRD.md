markdown
// filepath: /Users/arnut.k/Documents/workshops/gen-ai/workshop-be/PRD.md
# Product Requirements Document (PRD)
ระบบ Backend สำหรับ Authentication + Documentation + SQLite (Go Fiber)

## Goals
- ให้บริการ REST API สำหรับสมัครสมาชิก (Register) และเข้าสู่ระบบ (Login)
- ออก JWT Token สำหรับใช้งานใน endpoint ที่ต้อง authentication
- ใช้ SQLite เป็นฐานข้อมูล (local dev) พร้อมโครงสร้างรองรับการขยาย
- มีเอกสาร Swagger ครอบคลุมทุก route
- ปฏิบัติตามแนวทางการจัดเก็บข้อมูลที่ปลอดภัย (รหัสผ่านแฮช, ไม่เก็บ plaintext)

## Out of Scope (เวอร์ชันแรก)
- Refresh token / token rotation
- Role-based access control (RBAC)
- Social login
- Rate limiting
- Password reset flow

---

## 1. Database

### 1.1 Engine
- ใช้ SQLite (ไฟล์: data/app.db)
- เหตุผล: เบา, เหมาะกับ workshop / prototype, ไม่ต้อง setup server
- ใช้ไลบรารี: `gorm.io/gorm` + `gorm.io/driver/sqlite`

### 1.2 Schema (Initial)

Table: users
- id (uint, PK, auto increment)
- email (string, unique, not null, indexed)
- password_hash (string, not null)
- created_at (datetime)
- updated_at (datetime)
- last_login_at (nullable datetime)
- is_active (boolean, default true)

(ไม่เก็บ: plaintext password, ไม่เก็บ salt แยก ถ้าใช้ bcrypt ซึ่งจัดการภายใน)

### 1.3 การเชื่อมต่อ
- เปิด connection ตอน start
- ตรวจสอบว่าไฟล์โฟลเดอร์ data/ มีอยู่ (หากไม่มีก็สร้าง)
- Auto-migrate ด้วย GORM

### 1.4 Security / Compliance Notes
- Password: ใช้ bcrypt (cost 12 หรือ >= default) หรือ argon2id (ถ้าเพิ่ม lib)
- ห้าม log ค่า password / hash
- ใช้ parameterized execution (GORM จัดการ)
- ป้องกัน enumeration: ข้อความ error login ควรเป็น generic เช่น "invalid credentials"
- ปิด debug SQL ใน production

### 1.5 Migration Strategy
- เวอร์ชันแรกใช้ GORM AutoMigrate
- หากขยาย: เพิ่มเครื่องมือ migration เช่น `golang-migrate` ภายหลัง

---

## 2. Authentication

### 2.1 Flow
1) Register: ผู้ใช้ส่ง email + password
2) ระบบตรวจสอบว่าไม่ซ้ำ → แฮชรหัสผ่าน → บันทึก
3) Login: ผู้ใช้ส่ง email + password
4) ตรวจสอบ hash → ออก JWT access token
5) Endpoint ที่ต้องป้องกันใช้ Bearer token

### 2.2 Password Policy (เบื้องต้น)
- ความยาวขั้นต่ำ: 8 ตัวอักษร
- แนะนำให้ client ตรวจรูปแบบ (ตัวพิมพ์เล็ก/ใหญ่/ตัวเลข) แต่ไม่บังคับในเวอร์ชันแรก

### 2.3 JWT
- ใช้ไลบรารี: `github.com/golang-jwt/jwt/v5`
- Algorithm: HS256
- Secret: จาก ENV variable: JWT_SECRET (ถ้าไม่มีให้ panic)
- Expiration: 15 นาที (access token)
- Claims:
  - sub: user id
  - email
  - exp
  - iat
  - iss: "workshop-be"
- (Optional ภายหลัง: jti สำหรับ revoke / audit)

### 2.4 Error Handling Standard
- 400: validation error
- 401: invalid credentials / missing token
- 409: email already exists
- 500: internal error
Response format:
{
  "error": {
    "code": "EMAIL_EXISTS",
    "message": "email already registered"
  }
}

### 2.5 Rate Limiting (Future)
- (ระบุเป็น backlog)

---

## 3. API Endpoints

### 3.1 Health
GET /healthz
200 OK → purely for liveness

### 3.2 Register
POST /api/v1/auth/register
Request:
{
  "email": "user@example.com",
  "password": "P@ssw0rd123"
}
Responses:
- 201:
  {
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-09-18T12:00:00Z"
  }
- 409: email exists
- 400: invalid payload

### 3.3 Login
POST /api/v1/auth/login
Request:
{
  "email": "user@example.com",
  "password": "P@ssw0rd123"
}
Responses:
- 200:
  {
    "access_token": "<jwt>",
    "token_type": "Bearer",
    "expires_in": 900
  }
- 401: invalid credentials

### 3.4 Me (Protected Example)
GET /api/v1/auth/me
Header: Authorization: Bearer <token>
200:
{
  "id": 1,
  "email": "user@example.com",
  "last_login_at": "2025-09-18T13:20:00Z"
}

---

## 4. Documentation (Swagger / OpenAPI)

### 4.1 Tools
- ใช้ `github.com/gofiber/swagger` + `swag` (github.com/swaggo/swag/cmd/swag)
- คำสั่ง generate: `swag init -g main.go -o internal/docs`
- เส้นทาง Swagger UI: GET /swagger/index.html
- ไฟล์ที่ต้องมี: docs swagger (auto-generated) ไม่แก้ไขตรง

### 4.2 Annotation ตัวอย่าง
ใน handler / main.go:
 // @Summary Register user
 // @Tags Auth
 // @Accept json
 // @Produce json
 // @Param request body RegisterRequest true "register"
 // @Success 201 {object} RegisterResponse
 // @Failure 400 {object} ErrorResponse
 // @Failure 409 {object} ErrorResponse
 // @Router /api/v1/auth/register [post]

### 4.3 Definition Objects
- RegisterRequest
- RegisterResponse
- LoginRequest
- LoginResponse
- ErrorResponse
- UserMeResponse

---

## 5. Non-Functional Requirements

### 5.1 Performance
- รองรับ 50 RPS (workshop ใช้เครื่อง local)
- Response time เฉลี่ย < 200ms

### 5.2 Logging
- ใช้ log มาตรฐาน `log.Printf`
- ไม่ log password / token แบบเต็ม (สามารถ log prefix 5 ตัวแรกของ token ถ้าจำเป็น)

### 5.3 Configuration
ENV variables:
- PORT (default 3000)
- JWT_SECRET (ต้องกำหนด, ถ้าไม่มีก็ panic)
- APP_ENV (dev/prod) → ใช้กำหนด debug mode
- DB_PATH (default data/app.db)

### 5.4 Folder Structure (Proposed)
.
├── main.go
├── internal/
│   ├── db/
│   │   └── db.go
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── model.go
│   │   └── jwt.go
│   ├── middleware/
│   │   └── auth.go
│   └── docs/ (generated)
├── pkg/
│   └── password/ (แยก hash logic)
├── data/ (sqlite file)
├── PRD.md
├── go.mod

---

## 6. Validation Rules

Register:
- email: required, รูปแบบต้อง valid RFC5322 (ใช้ regex)
- password: required, length >= 8
Login:
- email + password required

---

## 7. Test Cases (High Level)

### 7.1 Register
- Success (ใหม่)
- Duplicate email
- Invalid email format
- Password too short

### 7.2 Login
- Success
- Email not found
- Wrong password

### 7.3 Auth Protected
- Missing token
- Invalid token signature
- Expired token

---

## 8. Future Enhancements (Backlog)
- Refresh token / rotation
- Password reset (email OTP)
- Account lockout (brute force defense)
- Soft delete users
- Audit log
- Prometheus metrics
- Docker + Compose (SQLite volume)

---

## 9. Acceptance Criteria (MVP)
- สามารถ register + login ได้
- ได้ JWT และ decode แล้วถูกต้อง
- /api/v1/auth/me ใช้ token แล้วได้ข้อมูลผู้ใช้
- Swagger UI แสดงทุก endpoint พร้อม schema
- ข้อมูลถูกเก็บในไฟล์ SQLite
- Password ไม่ถูกเก็บเป็น plaintext
- โค้ดรันผ่าน: go run main.go แล้วใช้งานตามที่ระบุ

---

## 10. Risks
- JWT secret เผลอ commit → ใช้ .env + .gitignore
- Race condition ตอน migrate (ต่ำมากใน single instance)
- Token ไม่ revoke ได้ (ยอมรับสำหรับ MVP)

---

## 11. Definition of Done
- โค้ด merge เข้าสาขาหลัก
- README อธิบายการรัน + swagger usage
- Swagger สร้างสำเร็จ
- Manual test ทุกกรณีหลักผ่าน
- ไม่มี secret ใน repo

---