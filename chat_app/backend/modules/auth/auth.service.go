package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"chat_app/backend/models"
	"chat_app/backend/modules/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthService interface {
	Register(input RegisterInput) (*models.User, string, error)
	Login(input LoginInput) (*models.User, string, error)
	GetUserByID(id uint) (*models.User, error)
	GenerateToken(user *models.User) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error)
}

type authService struct {
	userRepo  user.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo user.UserRepository, jwtSecret string) AuthService {
	if jwtSecret == "" {
		jwtSecret = "default_jwt_secret_key_chat_app"
	}
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Register(input RegisterInput) (*models.User, string, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Username == "" {
		return nil, "", errors.New("username is required")
	}
	if input.Email == "" {
		return nil, "", errors.New("email is required")
	}
	if len(input.Password) < 6 {
		return nil, "", errors.New("password must be at least 6 characters long")
	}

	// Check existing email
	existingEmailUser, err := s.userRepo.FindByEmail(input.Email)
	if err == nil && existingEmailUser != nil {
		return nil, "", errors.New("email is already registered")
	}

	// Check existing username
	existingUsernameUser, err := s.userRepo.FindByUsername(input.Username)
	if err == nil && existingUsernameUser != nil {
		return nil, "", errors.New("username is already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *authService) Login(input LoginInput) (*models.User, string, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email == "" || input.Password == "" {
		return nil, "", errors.New("email and password are required")
	}

	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil || user == nil {
		return nil, "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
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
