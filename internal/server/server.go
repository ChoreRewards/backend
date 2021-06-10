package server

import (
	"context"

	chorerewardsv1alpha1 "github.com/chorerewards/api/chorerewards/v1alpha1"
	"github.com/chorerewards/backend/internal/auth"
	"github.com/chorerewards/backend/internal/db"
	"github.com/pkg/errors"
)

// Server is the implementation of the chorerewardsv1alpha1.ChoreRewardsServiceServer
type Server struct {
	dbManager *db.Manager
}

type Config struct {
	DBHost     string
	DBPort     int
	DBUsername string
	DBPassword string
	DBName     string
}

func New(c Config) (*Server, error) {
	dbManager, err := db.New(db.Config{
		Host:     c.DBHost,
		Port:     c.DBPort,
		Username: c.DBUsername,
		Password: c.DBPassword,
		Database: c.DBName,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		dbManager: dbManager,
	}, nil
}

func (s *Server) ListUsers(context.Context, *chorerewardsv1alpha1.ListUsersRequest) (*chorerewardsv1alpha1.ListUsersResponse, error) {
	return nil, nil
}

func (s *Server) ListCategories(context.Context, *chorerewardsv1alpha1.ListCategoriesRequest) (*chorerewardsv1alpha1.ListCategoriesResponse, error) {
	return nil, nil
}

func (s *Server) ListTasks(context.Context, *chorerewardsv1alpha1.ListTasksRequest) (*chorerewardsv1alpha1.ListTasksResponse, error) {
	return nil, nil
}

func (s *Server) ListTasksFeed(context.Context, *chorerewardsv1alpha1.ListTasksFeedRequest) (*chorerewardsv1alpha1.ListTasksFeedResponse, error) {
	return nil, nil
}

func (s *Server) CreateUser(ctx context.Context, req *chorerewardsv1alpha1.CreateUserRequest) (*chorerewardsv1alpha1.CreateUserResponse, error) {
	pwdHash, err := auth.HashPassword([]byte(req.GetUser().GetPassword()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to hash password")
	}

	pinHash, err := auth.HashPassword([]byte(string(req.GetUser().GetPin())))
	if err != nil {
		return nil, errors.Wrap(err, "unable to hash pin")
	}

	user, err := s.dbManager.CreateUser(ctx, db.User{
		Username: req.GetUser().GetUsername(),
		Email:    req.GetUser().GetEmail(),
		IsAdmin:  req.GetUser().GetIsAdmin(),
		IsParent: req.GetUser().GetIsParent(),
		Avatar:   req.GetUser().GetAvatar(),
		Password: string(pwdHash),
		Pin:      string(pinHash),
		IsActive: true,
	})
	if err != nil {
		return nil, err
	}

	return &chorerewardsv1alpha1.CreateUserResponse{
		User: &chorerewardsv1alpha1.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
			IsParent: user.IsParent,
			Avatar:   user.Avatar,
			Points:   user.Points,
			IsActive: user.IsActive,
		},
	}, nil
}

func (s *Server) CreateCategory(context.Context, *chorerewardsv1alpha1.CreateCategoryRequest) (*chorerewardsv1alpha1.CreateCategoryResponse, error) {
	return nil, nil
}

func (s *Server) CreateTask(context.Context, *chorerewardsv1alpha1.CreateTaskRequest) (*chorerewardsv1alpha1.CreateTaskResponse, error) {
	return nil, nil
}

func (s *Server) AddTaskToFeed(context.Context, *chorerewardsv1alpha1.AddTaskToFeedRequest) (*chorerewardsv1alpha1.AddTaskToFeedResponse, error) {
	return nil, nil
}

func (s *Server) Login(ctx context.Context, req *chorerewardsv1alpha1.LoginRequest) (*chorerewardsv1alpha1.LoginResponse, error) {
	if req.GetUsername() == "" {
		return nil, errors.New("Username cannot be empty")
	}

	if req.GetPin() != 0 && req.GetPassword() != "" {
		return nil, errors.New("Specify either Pin OR Password, not both")
	}

	if req.GetPin() == 0 && req.GetPassword() == "" {
		return nil, errors.New("Specify either Pin OR Password")
	}

	user, err := s.dbManager.GetUser(ctx, req.GetUsername())
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get user")
	}

	var authenticated bool

	if req.GetPin() != 0 {
		authenticated = auth.PasswordMatches([]byte(user.Pin), []byte(string(req.GetPin())))
	} else {
		authenticated = auth.PasswordMatches([]byte(user.Password), []byte(req.GetPassword()))
	}

	if !authenticated {
		return nil, errors.New("Authentication failed")
	}

	return &chorerewardsv1alpha1.LoginResponse{}, nil
}
