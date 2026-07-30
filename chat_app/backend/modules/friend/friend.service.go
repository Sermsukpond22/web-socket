package friend

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"chat_app/backend/models"
	"chat_app/backend/modules/user"

	"gorm.io/gorm"
)

type SendRequestInput struct {
	ToUserID   uint   `json:"to_user_id"`
	ToUsername string `json:"to_username"`
}

type FriendService interface {
	SendFriendRequest(fromUserID uint, target string) (*models.FriendRequest, error)
	SendFriendRequestByInput(fromUserID uint, input SendRequestInput) (*models.FriendRequest, error)
	GetPendingRequests(userID uint) ([]models.FriendRequest, error)
	AcceptFriendRequest(requestID, userID uint) (*models.FriendRequest, error)
	GetFriends(userID uint) ([]models.User, error)
	AreFriends(userAID, userBID uint) (bool, error)
	SearchUsers(query string, excludeUserID uint) ([]models.User, error)
}

type friendService struct {
	friendRepo FriendRepository
	userRepo   user.UserRepository
}

func NewFriendService(friendRepo FriendRepository, userRepo user.UserRepository) FriendService {
	return &friendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
	}
}

func (s *friendService) SendFriendRequest(fromUserID uint, target string) (*models.FriendRequest, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, errors.New("target user must be specified")
	}

	var targetUser *models.User
	var err error

	// Try lookup by username first
	targetUser, err = s.userRepo.FindByUsername(target)
	if err != nil || targetUser == nil {
		// Try parsing as user ID
		if parsedID, parseErr := strconv.ParseUint(target, 10, 64); parseErr == nil {
			targetUser, err = s.userRepo.FindByID(uint(parsedID))
		}
	}

	if err != nil || targetUser == nil {
		return nil, errors.New("user not found")
	}

	return s.processSendRequest(fromUserID, targetUser)
}

func (s *friendService) SendFriendRequestByInput(fromUserID uint, input SendRequestInput) (*models.FriendRequest, error) {
	var targetUser *models.User
	var err error

	if input.ToUserID > 0 {
		targetUser, err = s.userRepo.FindByID(input.ToUserID)
	} else if strings.TrimSpace(input.ToUsername) != "" {
		targetUser, err = s.userRepo.FindByUsername(strings.TrimSpace(input.ToUsername))
	} else {
		return nil, errors.New("target user must be specified via to_user_id or to_username")
	}

	if err != nil || targetUser == nil {
		return nil, errors.New("user not found")
	}

	return s.processSendRequest(fromUserID, targetUser)
}

func (s *friendService) processSendRequest(fromUserID uint, targetUser *models.User) (*models.FriendRequest, error) {
	if fromUserID == targetUser.ID {
		return nil, errors.New("cannot send friend request to yourself")
	}

	// Check if already friends
	areFriends, err := s.friendRepo.AreFriends(fromUserID, targetUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check friendship status: %w", err)
	}
	if areFriends {
		return nil, errors.New("users are already friends")
	}

	// Check existing requests
	existingReq, err := s.friendRepo.FindRequestBetweenUsers(fromUserID, targetUser.ID)
	if err == nil && existingReq != nil {
		if existingReq.Status == "accepted" {
			return nil, errors.New("users are already friends")
		}
		if existingReq.Status == "pending" {
			if existingReq.FromUserID == fromUserID {
				return nil, errors.New("friend request already sent")
			}
			return nil, errors.New("a pending friend request from this user already exists")
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Log or handle non-RecordNotFound database errors
	}

	return s.friendRepo.SendRequest(fromUserID, targetUser.ID)
}

func (s *friendService) GetPendingRequests(userID uint) ([]models.FriendRequest, error) {
	return s.friendRepo.GetPendingRequests(userID)
}

func (s *friendService) AcceptFriendRequest(requestID, userID uint) (*models.FriendRequest, error) {
	req, err := s.friendRepo.FindRequestByID(requestID)
	if err != nil || req == nil {
		return nil, errors.New("friend request not found")
	}

	if req.ToUserID != userID {
		return nil, errors.New("unauthorized to accept this request")
	}

	if req.Status != "pending" {
		return nil, errors.New("friend request is not pending")
	}

	acceptedReq, err := s.friendRepo.AcceptRequest(requestID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.friendRepo.CreateFriendship(req.FromUserID, req.ToUserID); err != nil {
		return nil, fmt.Errorf("failed to create friendship: %w", err)
	}

	return acceptedReq, nil
}

func (s *friendService) GetFriends(userID uint) ([]models.User, error) {
	return s.friendRepo.GetFriendsList(userID)
}

func (s *friendService) AreFriends(userAID, userBID uint) (bool, error) {
	return s.friendRepo.AreFriends(userAID, userBID)
}

func (s *friendService) SearchUsers(query string, excludeUserID uint) ([]models.User, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []models.User{}, nil
	}
	return s.userRepo.SearchUsers(query, excludeUserID)
}
