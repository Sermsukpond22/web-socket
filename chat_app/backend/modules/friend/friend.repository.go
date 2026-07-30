package friend

import (
	"errors"

	"chat_app/backend/models"

	"gorm.io/gorm"
)

type FriendRepository interface {
	SendRequest(fromUserID, toUserID uint) (*models.FriendRequest, error)
	GetPendingRequests(userID uint) ([]models.FriendRequest, error)
	FindRequestByID(requestID uint) (*models.FriendRequest, error)
	FindRequestBetweenUsers(userAID, userBID uint) (*models.FriendRequest, error)
	AcceptRequest(requestID, userID uint) (*models.FriendRequest, error)
	CreateFriendship(userAID, userBID uint) error
	AreFriends(userAID, userBID uint) (bool, error)
	GetFriendsList(userID uint) ([]models.User, error)
}

type friendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{db: db}
}

func (r *friendRepository) SendRequest(fromUserID, toUserID uint) (*models.FriendRequest, error) {
	req := &models.FriendRequest{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Status:     "pending",
	}
	if err := r.db.Create(req).Error; err != nil {
		return nil, err
	}
	if err := r.db.Preload("FromUser").Preload("ToUser").First(req, req.ID).Error; err != nil {
		return nil, err
	}
	return req, nil
}

func (r *friendRepository) GetPendingRequests(userID uint) ([]models.FriendRequest, error) {
	var requests []models.FriendRequest
	err := r.db.Preload("FromUser").Preload("ToUser").
		Where("to_user_id = ? AND status = ?", userID, "pending").
		Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *friendRepository) FindRequestByID(requestID uint) (*models.FriendRequest, error) {
	var req models.FriendRequest
	err := r.db.Preload("FromUser").Preload("ToUser").First(&req, requestID).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *friendRepository) FindRequestBetweenUsers(userAID, userBID uint) (*models.FriendRequest, error) {
	var req models.FriendRequest
	err := r.db.Preload("FromUser").Preload("ToUser").
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			userAID, userBID, userBID, userAID).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *friendRepository) AcceptRequest(requestID, userID uint) (*models.FriendRequest, error) {
	req, err := r.FindRequestByID(requestID)
	if err != nil {
		return nil, err
	}
	if req.ToUserID != userID {
		return nil, errors.New("unauthorized to accept this request")
	}
	if req.Status != "pending" {
		return nil, errors.New("request is not pending")
	}

	req.Status = "accepted"
	if err := r.db.Model(req).Update("status", "accepted").Error; err != nil {
		return nil, err
	}

	return req, nil
}

func (r *friendRepository) CreateFriendship(userAID, userBID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var f1 models.Friendship
		if err := tx.Where("user_id = ? AND friend_id = ?", userAID, userBID).First(&f1).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&models.Friendship{UserID: userAID, FriendID: userBID}).Error; err != nil {
				return err
			}
		}

		var f2 models.Friendship
		if err := tx.Where("user_id = ? AND friend_id = ?", userBID, userAID).First(&f2).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&models.Friendship{UserID: userBID, FriendID: userAID}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *friendRepository) AreFriends(userAID, userBID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Friendship{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userAID, userBID, userBID, userAID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *friendRepository) GetFriendsList(userID uint) ([]models.User, error) {
	var friendships []models.Friendship
	err := r.db.Preload("Friend").Where("user_id = ?", userID).Find(&friendships).Error
	if err != nil {
		return nil, err
	}

	friends := make([]models.User, 0, len(friendships))
	for _, f := range friendships {
		friends = append(friends, f.Friend)
	}

	return friends, nil
}
