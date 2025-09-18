# Copilot / AI Assistant Project Instructions (workshop-be)

เอกสารนี้อธิบายวิธีสั่งงาน (prompt) AI / Copilot สำหรับโปรเจ็ค Go Fiber Backend นี้ ให้ผลลัพธ์ถูกต้อง ปลอดภัย สม่ำเสมอ และขยายต่อได้ง่าย

---
## 1. วัตถุประสงค์ระบบ
ระบบนี้ให้บริการ REST API สำหรับ:
- Register / Login และออก JWT (HS256)
- Endpoint ป้องกันด้วย Bearer Token (/api/v1/auth/me, /api/v1/profile)
- จัดเก็บข้อมูลผู้ใช้ + โปรไฟล์ใน SQLite (ผ่าน GORM)
- มีเอกสาร Swagger ครอบคลุม
- รองรับการขยาย (Profile, Membership, Points)

---
## 2. Tech Stack & Libraries
- Language: Go 1.22+ (ห้ามใช้ฟีเจอร์ experimental ที่ไม่เสถียร)
- Web Framework: Fiber
- ORM: GORM + SQLite driver
- Auth: github.com/golang-jwt/jwt/v5
- Password Hash: bcrypt (cost 12)
- Env: godotenv
- Docs: swag / gofiber/swagger
- Hot Reload: Air (.air.toml)

---
## 3. หลักการสำคัญ (Guiding Principles)
1. Security First: ไม่ log secret / token / password
2. Explicit Validation: ทุก input ผ่าน validation ก่อน persistence
3. Least Exposure: Read-only fields แยกชัดเจนใน profile update
4. Consistent Error Schema: `{"error":{"code":"...","message":"..."}}`
5. Deterministic Behavior: Partial update ไม่ทำให้ field อื่นเป็น zero value โดยไม่ตั้งใจ
6. No Silent Failures: ทุกกรณีผิดปกติส่ง error หรือ panic ที่ควร (เช่น JWT_SECRET ขาดใน prod)
7. Separation of Concerns: handler (transport) -> service (business) -> model (data)
8. Idempotency สำหรับ Register ไม่ทำ (แต่ error 409) – อย่า auto login หลัง register เวอร์ชันนี้
9. Observability Ready: logging ควร structured-friendly (prefix key=value)
10. Minimal Surface: อย่าเพิ่ม dependency โดยไม่จำเป็น

---
## 4. โครงสร้างโฟลเดอร์ (ปัจจุบัน / ขยาย)
```
internal/
  auth/ (model.go, service.go, handler.go, jwt.go)
  db/   (db.go init + migrations)
  middleware/ (auth.go)
  docs/ (generated – ห้ามแก้ตรง)
pkg/password/ (bcrypt utilities)
main.go
PRD.md
```
แนวทางขยาย: เพิ่ม package ใหม่ภายใต้ `internal/<domain>` แล้วคง pattern service + handler + model

---
## 5. สิ่งที่ AI ควรรักษา (Invariants)
- ห้ามแก้ไขไฟล์ swagger generated (`internal/docs/*`) ด้วยมือ ให้ใช้ `swag init`
- ห้าม log plaintext password หรือ hash
- JWT secret ต้องอ่านจาก ENV (prod: บังคับมี, dev: อนุญาต fallback พร้อมคำเตือน)
- Subject (sub) claim เป็น string ของ user ID (อย่าใช้ raw int)
- เวลาใช้ context: คีย์ใน middleware: `user_email`, `user_sub`
- Profile PUT: ห้ามแก้ membership_level / membership_code / points / joined_at
- Response JSON field names ต้องสอดคล้อง Swagger
- Error code ต้องใช้ enum ที่ประกาศไว้ (เช่น INVALID_PHONE, EMAIL_EXISTS)

---
## 6. Validation Rules (สรุปสำหรับ AI)
- Email: regex RFC5322 (มีแล้ว – อย่าเปลี่ยนรูปแบบโดยไม่มีเหตุผล)
- Password: length >= 8
- First/Last Name: ถ้ามี 1..100 chars (trim)
- Phone: ถ้ามี → เก็บเฉพาะ digits 10 หลัก (ไทย) มิฉะนั้น INVALID_PHONE

---
## 7. JWT & Auth
- Algorithm: HS256 เท่านั้น (อย่า fallback none / RS256 โดยพลการ)
- Exp: 15 นาที (อย่า hardcode อื่น เว้นเพิ่ม config)
- Claims: sub, email, exp, iat, iss="workshop-be"
- Middleware: ตรวจ Bearer, parse, ใส่ข้อมูลใน ctx.locals
- อย่า expose token decoding endpoint (ไม่จำเป็น)

---
## 8. Error Handling Pattern
ตัวอย่าง:
```
return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: ErrorObject{
  Code:    "INVALID_PHONE",
  Message: "phone must be 10 digits",
}})
```
ห้ามคืน error เปล่า / panic ใน handler (ยกเว้น panic ที่ตั้งใจตอน config ผิดใน main)

---
## 9. Logging Guidelines
- ใช้ log.Printf("event=login_success user_id=%d email=%s", user.ID, user.Email)
- Error: log.Printf("level=error event=login_failed reason=%s email=%s", err, req.Email)
- ห้าม log token ทั้งหมด (ถ้าจำเป็น แสดงเฉพาะ prefix 8 ตัวแรก + "…")

---
## 10. Swagger Annotations (Checklist ต่อ endpoint ใหม่)
1. @Summary (สั้น กระชับ)
2. @Description (ถ้ามี business rule พิเศษ)
3. @Tags <Domain>
4. @Accept json (ถ้ามี body)
5. @Produce json
6. @Security BearerAuth (ถ้าต้อง auth)
7. @Param request body <Type> true "description"
8. @Success <code> {object} <RespType>
9. @Failure <code> {object} ErrorResponse
10. @Router /path [method]
หลังแก้: รัน `swag init -g main.go -o internal/docs`

---
## 11. Testing Strategy (ที่จะเพิ่ม)
ประเภท:
- Unit: service logic (validation, hashing, profile normalization)
- Integration: register->login->me->profile flow (ใช้ temp SQLite :memory: หรือไฟล์ชั่วคราว)
แนวทางเขียน test:
- ใช้ table-driven tests
- แยก helper สร้าง user / token
Priority (ตาม PRD Test Cases): Register, Login, Auth middleware, Profile

Template ตัวอย่าง test case (pseudocode):
```
func TestRegister(t *testing.T) {
  cases := []struct{name string; email string; password string; wantErr string}{...}
  for _, tc := range cases { ... }
}
```

---
## 12. Security Checklist (ขอให้ AI ตรวจทุกครั้งก่อน PR)
- [ ] ไม่มี hardcoded secret
- [ ] JWT secret panic ถ้า APP_ENV=prod และค่าว่าง
- [ ] ไม่ log password/token
- [ ] Validation ครบ (email/password/profile)
- [ ] Error message ไม่เปิดเผยข้อมูลเกิน (เช่น login)
- [ ] Dependencies ไม่เพิ่มโดยไม่จำเป็น

---
## 13. Performance & Reliability Notes
- ไม่ทำ query N+1 (ปัจจุบันเรียก user เดี่ยว – OK)
- ควร reuse DB connection (มีแล้วใน db.go)
- เพิ่ม index (email, phone) – ให้ AI แจ้งหากยังไม่มี tag gorm ที่เหมาะสม

---
## 14. Git / PR Guidelines
Commit message format (แนะนำ):
```
feat(auth): add /auth/me endpoint
fix(profile): correct phone normalization edge case
chore(docs): regenerate swagger after profile changes
```
PR Checklist (AI ควรเตือน):
- [ ] Swagger updated
- [ ] README อัพเดตถ้า endpoint ใหม่
- [ ] Tests เพิ่ม/อัพเดต (ถ้าแตะ logic)
- [ ] ไม่มีไฟล์ generated ถูกแก้ด้วยมือ
- [ ] Run build & (ถ้าเพิ่ม) run tests ผ่าน

---
## 15. Prompt Patterns (ตัวอย่างสั่งงาน AI ให้ได้ผลดี)
ขอเพิ่มฟีเจอร์ใหม่:
"ช่วยเพิ่ม endpoint POST /api/v1/auth/refresh (ยังไม่ implement token rotation เต็ม ให้ mock response) พร้อม swagger และอัพเดต README"

ขอรีแฟคเตอร์:
"รีแฟคเตอร์ service สมัครสมาชิก แยก validation ออกเป็นฟังก์ชัน registerValidate() และเพิ่ม unit test"

ขอ test:
"เขียน unit test สำหรับฟังก์ชัน UpdateProfile ครอบคลุม: success, invalid phone, invalid name, partial update"

ตรวจ security:
"รีวิวไฟล์ service.go หา point ที่อาจ log sensitive data และเสนอแก้ไข"

---
## 16. Anti-Patterns (ห้ามทำ)
- สร้าง global mutable state นอกเหนือจาก DB connection / config
- Return interface{} แบบหลวม ๆ แทน struct ที่ชัดเจน
- ใช้ panic ภายใน business logic (ยกเว้น config fatal ที่ startup)
- Copy-paste handler logic แทนการแยก service method
- แก้ไขไฟล์ swagger ด้วยมือ
- เพิ่ม field ลง User โดยไม่อัพเดต PRD + Swagger
- ใช้เวลา current time โดยไม่ผ่าน time.Now() ตรง ๆ ใน test (ให้จัด abstraction เมื่อเริ่มเขียน test ชุดใหญ่)

---
## 17. การเพิ่ม Endpoint ใหม่ (Quick Checklist)
1. ออกแบบ request/response + validation
2. เพิ่ม struct ใน model/service
3. เขียน service logic + tests (ถ้ามี infra พร้อม)
4. เขียน handler + swagger annotation
5. ลงทะเบียน route ใน main.go
6. รัน `swag init` regenerate docs
7. เพิ่ม README section (ถ้า public endpoint ใหม่)
8. Manual curl ทดสอบ (บันทึกตัวอย่าง)
9. ตรวจ Security Checklist

---
## 18. Environment Variables & Config
Required:
- PORT (default 3000)
- APP_ENV (dev|prod) – prod จะ disable insecure fallback
- JWT_SECRET (prod ต้องมี ไม่งั้น panic)
- DB_PATH (default data/app.db)
ห้าม commit ไฟล์ .env ที่มี secret จริง (ใช้ .env.example สื่อรูปแบบ)

---
## 19. การจัดการ Membership / Points (Future Hooks)
เตรียม field แล้ว: membership_level, membership_code, points.
หากเพิ่ม endpoint:
- แยก domain ใหม่ (internal/membership/*)
- ใช้ transaction เมื่อมี points deduction/add
- เพิ่ม audit log table ภายหลัง (users_points_history)

---
## 20. Observability (Backlog Spec สำหรับ AI)
เมื่อ implement:
- เพิ่ม middleware timing (เริ่ม/จบ request)
- เพิ่ม metrics (Prometheus) path, status_code, latency_bucket
- อย่าบันทึก body ของ request ที่มี password

---
## 21. ตัวอย่าง Prompt "ไม่ดี" vs "ดี"
ไม่ดี: "เพิ่ม profile" (กว้างเกินไป)
ดี: "เพิ่ม unit test ให้ UpdateProfile ครอบคลุม invalid phone (มีตัวอักษร), invalid name (ค่าว่าง), partial update (first_name อย่างเดียว) ใช้ table-driven pattern"

---
## 22. การอัพเดต PRD.md
ทุกครั้งที่เพิ่ม domain / field ใหม่:
- อัพเดต Section Schema
- เพิ่ม Test Cases + Validation
- เพิ่ม Acceptance Criteria หากมีผลลัพธ์ใหม่ที่ต้องยืนยัน
- ถ้า flow ใหม่ เพิ่ม Mermaid diagram (sequence หรือ context)

---
## 23. Mermaid Diagrams (แนวทาง)
- เก็บต้นฉบับใน PRD.md พร้อมสามารถแยกไฟล์ .mmd ใน design/ ถ้าเริ่มเยอะ
- ควรมี: Context, Register Sequence, Login Sequence, Profile Get/Update Sequence

---
## 24. สคริปต์ / คำสั่ง Dev (AI ควรจำแนกเวลาอ้างอิง)
รัน dev (hot reload): `air`
Generate swagger: `swag init -g main.go -o internal/docs`
รันโปรแกรมตรง: `go run main.go`
รัน test (ภายหลัง): `go test ./... -count=1`

---
## 25. การจัดรูปแบบโค้ด
- ใช้ `gofmt` / `goimports`
- ชื่อไฟล์ lower_snake.go
- ชื่อ struct เป็น PascalCase, ชื่อ private method camelCase
- Error var เริ่มด้วย `Err` เช่น `ErrInvalidPhone`

---
## 26. Concurrency Considerations (ตอนขยาย)
- ปัจจุบันไม่ใช้ goroutine พิเศษใน business logic
- ถ้าเพิ่ม background job: ใช้ context cancel + graceful shutdown
- อย่า share *gorm.DB copy ในหลาย routine โดย mutate config runtime

---
## 27. Checklist ก่อน Merge (รวม)
- [ ] Build ผ่าน (go build ./...)
- [ ] Swagger regenerate แล้ว ไม่มี stale annotation
- [ ] README/PRD อัพเดต
- [ ] Security checklist ผ่าน
- [ ] ไม่มี debug print เหลือ
- [ ] Tests (ถ้ามี) ผ่าน `go test ./...`

---
## 28. คำแนะนำสำหรับ AI เวลาแก้ไฟล์
- อ่านไฟล์ก่อนแก้ (context)
- ใช้ pattern comment `// ...existing code...` เวลาส่ง patch (หากระบบต้องการ)
- อย่าลบ logic validation เดิมโดยไม่ระบุเหตุผล
- ถ้ามีการ refactor: ระบุผลกระทบ swagger/test

---
## 29. การตั้งชื่อ Error Codes ใหม่
Format: UPPER_SNAKE, สั้น, สื่อชัด (เช่น INVALID_EMAIL_FORMAT, DUPLICATE_PHONE)
อย่าใช้ข้อความทั่วไปเช่น ERROR_1

---
## 30. ความแตกต่าง Dev vs Prod (AI ต้องรักษา)
Dev:
- อนุญาต fallback JWT secret (พร้อม log warning เด่น)
Prod:
- ถ้าไม่มี JWT_SECRET ให้ panic ทันที (TODO implement – เพิ่ม test ภายหลัง)
- ปิด verbose log / debug

---
## 31. Roadmap (Short)
- เพิ่ม test suite
- Enforce prod secret requirement
- Structured logging (zap หรือ zerolog) – ต้องปรึกษาก่อนเพิ่ม lib
- Dockerfile + docker-compose
- Refresh token flow

---
## 32. วิธีถาม AI เพื่อรีวิวโค้ด
ตัวอย่าง: "รีวิว internal/auth/service.go หา logic duplication และเสนอ refactor เป็น helper พร้อมอธิบาย trade-off"

---
## 33. สรุป
ใช้เอกสารนี้เป็นมาตรฐานทุกครั้งที่:
- เพิ่ม endpoint
- แก้ business rule
- ขอให้ AI เขียน test / refactor

หากข้อเสนอของ AI ขัดกับ invariant ให้ปฏิเสธ / ขอเหตุผลเชิงเทคนิคก่อนดำเนินการ

---
(End of Instructions)