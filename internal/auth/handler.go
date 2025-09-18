package auth

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeError(c *fiber.Ctx, status int, code, msg string) error {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = msg
	return c.Status(status).JSON(resp)
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register godoc
// @Summary Register user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterInput true "register"
// @Success 201 {object} RegisterOutput
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *fiber.Ctx) error {
	var in RegisterInput
	if err := c.BodyParser(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid payload")
	}
	out, err := h.svc.Register(in)
	if err != nil {
		switch err {
		case ErrInvalidEmail:
			return writeError(c, http.StatusBadRequest, "INVALID_EMAIL", "invalid email")
		case ErrPasswordTooShort:
			return writeError(c, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "password too short")
		case ErrEmailExists:
			return writeError(c, http.StatusConflict, "EMAIL_EXISTS", "email already registered")
		default:
			return writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
	}
	return c.Status(http.StatusCreated).JSON(out)
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginInput true "login"
// @Success 200 {object} LoginOutput
// @Failure 401 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *fiber.Ctx) error {
	var in LoginInput
	if err := c.BodyParser(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid payload")
	}
	out, err := h.svc.Login(in)
	if err != nil {
		if err == ErrInvalidCredential {
			return writeError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials")
		}
		return writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}
	return c.JSON(out)
}

func RegisterRoutes(r fiber.Router, svc *Service) {
	h := NewHandler(svc)
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
}
