package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"chat_app/backend/models"
	"chat_app/backend/modules/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthService interface {
	Register(input RegisterInput) (*models.User, error)
	Login(input LoginInput) (*models.User, string, string, error)
	GetUserByID(id uint) (*models.User, error)
	GenerateToken(user *models.User) (string, error)
	GenerateRefreshToken(user *models.User) (string, error)
	ValidateRefreshToken(tokenString string) (*jwt.Token, jwt.MapClaims, error)
	ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error)
}

type authService struct {
	userRepo  user.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo user.UserRepository, jwtSecret string) AuthService {

	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Register(input RegisterInput) (*models.User, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Bio = strings.TrimSpace(input.Bio)

	if input.Username == "" {
		return nil, errors.New("กรุณากรอกชื่อผู้ใช้ (Username)")
	}
	if input.Email == "" {
		return nil, errors.New("กรุณากรอกอีเมล")
	}
	if !strings.Contains(input.Email, "@") {
		return nil, errors.New("รูปแบบอีเมลไม่ถูกต้อง")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("รหัสผ่านต้องมีความยาวอย่างน้อย 6 ตัวอักษร")
	}
	if input.DisplayName == "" {
		return nil, errors.New("กรุณากรอกชื่อแสดงผล (Display Name)")
	}

	// Check existing email
	existingEmailUser, err := s.userRepo.FindByEmail(input.Email)
	if err == nil && existingEmailUser != nil {
		return nil, errors.New("อีเมลนี้ถูกใช้งานแล้ว")
	}

	// Check existing username
	existingUsernameUser, err := s.userRepo.FindByUsername(input.Username)
	if err == nil && existingUsernameUser != nil {
		return nil, errors.New("ชื่อผู้ใช้นี้ถูกใช้งานแล้ว")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("เข้ารหัสผ่านไม่สำเร็จ: %w", err)
	}

	user := &models.User{
		Username:    input.Username,
		Email:       input.Email,
		Password:    string(hashedPassword),
		DisplayName: input.DisplayName,
		Bio:         input.Bio,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("ไม่สามารถสร้างบัญชีผู้ใช้ได้: %w", err)
	}

	return user, nil
}

func (s *authService) Login(input LoginInput) (*models.User, string, string, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email == "" || input.Password == "" {
		return nil, "", "", errors.New("email and password are required")
	}

	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil || user == nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, token, refreshToken, nil

}

func (s *authService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *authService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, errors.New("invalid token")
}

func (s *authService) GenerateRefreshToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = s.jwtSecret // fallback
	}
	return token.SignedString([]byte(refreshSecret))
}

func (s *authService) ValidateRefreshToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = s.jwtSecret // fallback
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(refreshSecret), nil
	})

	if err != nil {
		return nil, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, errors.New("invalid refresh token")
}
