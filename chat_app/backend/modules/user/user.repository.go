package user

import (
	"strings"
	"time"

	"chat_app/backend/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	SearchUsers(query string, excludeUserID uint) ([]models.User, error)
	UpdateLastSeen(userID uint, lastSeen *time.Time) error
	UpdateProfile(userID uint, displayName string, bio string) error
	UpdateAvatar(userID uint, avatarURL string) error
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
	escapedQuery := strings.ReplaceAll(trimmed, "%", "\\%")
	escapedQuery = strings.ReplaceAll(escapedQuery, "_", "\\_")
	likeQuery := "%" + strings.ToLower(escapedQuery) + "%"
	err := r.db.Where("id != ? AND (LOWER(username) LIKE ? OR LOWER(email) LIKE ?)", excludeUserID, likeQuery, likeQuery).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) UpdateLastSeen(userID uint, lastSeen *time.Time) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("last_seen", lastSeen).Error
}

func (r *userRepository) UpdateProfile(userID uint, displayName string, bio string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(models.User{
		DisplayName: displayName,
		Bio:         bio,
	}).Error
}

func (r *userRepository) UpdateAvatar(userID uint, avatarURL string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error
}
