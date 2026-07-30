package user

import (
	"strings"

	"chat_app/backend/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	SearchUsers(query string, excludeUserID uint) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) SearchUsers(query string, excludeUserID uint) ([]models.User, error) {
	var users []models.User
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []models.User{}, nil
	}
	likeQuery := "%" + strings.ToLower(trimmed) + "%"
	err := r.db.Where("id != ? AND (LOWER(username) LIKE ? OR LOWER(email) LIKE ?)", excludeUserID, likeQuery, likeQuery).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
