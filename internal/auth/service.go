package auth

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"workshop-be/internal/db"
	"workshop-be/pkg/password"
)

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
var phoneDigitsRegex = regexp.MustCompile(`\D+`)

var (
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrInvalidPhone      = errors.New("invalid phone")
	ErrInvalidName       = errors.New("invalid name")
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

// Profile types

type ProfileResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	FirstName       *string    `json:"first_name"`
	LastName        *string    `json:"last_name"`
	Phone           *string    `json:"phone"`
	MembershipLevel string     `json:"membership_level"`
	MembershipCode  *string    `json:"membership_code"`
	Points          int        `json:"points"`
	JoinedAt        *time.Time `json:"joined_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ProfileUpdateRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
}

type MeOutput struct {
	ID          uint       `json:"id"`
	Email       string     `json:"email"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

func (s *Service) Register(input RegisterInput) (*RegisterOutput, error) {
	if !emailRegex.MatchString(input.Email) {
		return nil, ErrInvalidEmail
	}
	if len(input.Password) < 8 {
		return nil, ErrPasswordTooShort
	}
	d := db.MustGet()
	var count int64
	d.Model(&User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		return nil, ErrEmailExists
	}
	h, err := password.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	user := User{Email: input.Email, PasswordHash: h}
	if err := d.Create(&user).Error; err != nil {
		return nil, err
	}
	return &RegisterOutput{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt}, nil
}

func (s *Service) Login(input LoginInput) (*LoginOutput, error) {
	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidCredential
	}
	d := db.MustGet()
	var user User
	if err := d.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return nil, ErrInvalidCredential
	}
	if !password.Verify(user.PasswordHash, input.Password) {
		return nil, ErrInvalidCredential
	}
	now := time.Now()
	d.Model(&user).Update("last_login_at", &now)
	expiry := 15 * time.Minute
	token, err := GenerateToken(user.ID, user.Email, expiry)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(expiry.Seconds())}, nil
}

// GetProfile returns user profile by id
func (s *Service) GetProfile(userID uint) (*ProfileResponse, error) {
	d := db.MustGet()
	var user User
	if err := d.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:              user.ID,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Phone:           user.Phone,
		MembershipLevel: user.MembershipLevel,
		MembershipCode:  user.MembershipCode,
		Points:          user.Points,
		JoinedAt:        user.JoinedAt,
		CreatedAt:       user.CreatedAt,
	}, nil
}

// UpdateProfile updates editable fields
func (s *Service) UpdateProfile(userID uint, req ProfileUpdateRequest) (*ProfileResponse, error) {
	d := db.MustGet()
	var user User
	if err := d.First(&user, userID).Error; err != nil {
		return nil, err
	}
	// Validate names
	if req.FirstName != nil {
		fn := strings.TrimSpace(*req.FirstName)
		if len(fn) == 0 || len(fn) > 100 {
			return nil, ErrInvalidName
		}
		user.FirstName = &fn
	}
	if req.LastName != nil {
		ln := strings.TrimSpace(*req.LastName)
		if len(ln) == 0 || len(ln) > 100 {
			return nil, ErrInvalidName
		}
		user.LastName = &ln
	}
	if req.Phone != nil {
		processed := phoneDigitsRegex.ReplaceAllString(*req.Phone, "")
		if len(processed) != 10 {
			return nil, ErrInvalidPhone
		}
		user.Phone = &processed
	}
	if err := d.Save(&user).Error; err != nil {
		return nil, err
	}
	return s.GetProfile(user.ID)
}

func (s *Service) Me(userID uint) (*MeOutput, error) {
	d := db.MustGet()
	var u User
	if err := d.First(&u, userID).Error; err != nil {
		return nil, err
	}
	return &MeOutput{ID: u.ID, Email: u.Email, LastLoginAt: u.LastLoginAt}, nil
}
