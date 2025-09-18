package auth

import (
	"errors"
	"regexp"
	"time"

	"workshop-be/internal/db"
	"workshop-be/pkg/password"
)

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

var (
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrInvalidCredential = errors.New("invalid credentials")
)

type Service struct{}

func NewService() *Service { return &Service{} }

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterOutput struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (s *Service) Register(input RegisterInput) (*RegisterOutput, error) {
	if !emailRegex.MatchString(input.Email) { return nil, ErrInvalidEmail }
	if len(input.Password) < 8 { return nil, ErrPasswordTooShort }
	d := db.MustGet()
	var count int64
	d.Model(&User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 { return nil, ErrEmailExists }
	h, err := password.Hash(input.Password)
	if err != nil { return nil, err }
	user := User{Email: input.Email, PasswordHash: h}
	if err := d.Create(&user).Error; err != nil { return nil, err }
	return &RegisterOutput{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt}, nil
}

func (s *Service) Login(input LoginInput) (*LoginOutput, error) {
	if input.Email == "" || input.Password == "" { return nil, ErrInvalidCredential }
	d := db.MustGet()
	var user User
	if err := d.Where("email = ?", input.Email).First(&user).Error; err != nil { return nil, ErrInvalidCredential }
	if !password.Verify(user.PasswordHash, input.Password) { return nil, ErrInvalidCredential }
	now := time.Now()
	d.Model(&user).Update("last_login_at", &now)
	expiry := 15 * time.Minute
	token, err := GenerateToken(user.ID, user.Email, expiry)
	if err != nil { return nil, err }
	return &LoginOutput{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(expiry.Seconds())}, nil
}
