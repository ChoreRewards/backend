package server

import (
	"context"

	"github.com/chorerewards/backend/internal/auth"
	"github.com/chorerewards/backend/internal/db"
	chorerewardsv1alpha1 "github.com/chorerewards/proto/chorerewards/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errNotFound *db.ErrNotFound

type TokenManager interface {
	CreateToken(username string) (string, error)
}

// Server is the implementation of the chorerewardsv1alpha1.ChoreRewardsServiceServer
type Server struct {
	dbManager    *db.Manager
	tokenManager TokenManager
}

type Config struct {
	DBHost     string
	DBPort     int
	DBUsername string
	DBPassword string
	DBName     string
}

func New(c Config, tokenManager TokenManager) (*Server, error) {
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
		dbManager:    dbManager,
		tokenManager: tokenManager,
	}, nil
}

func (s *Server) CreateCategory(ctx context.Context, req *chorerewardsv1alpha1.CreateCategoryRequest) (*chorerewardsv1alpha1.CreateCategoryResponse, error) {
	category, err := s.dbManager.CreateCategory(ctx, db.Category{
		Color:       req.GetCategory().GetColor(),
		Name:        req.GetCategory().GetName(),
		Description: req.GetCategory().GetDescription(),
	})
	if err != nil {
		return nil, err
	}

	return &chorerewardsv1alpha1.CreateCategoryResponse{
		Category: &chorerewardsv1alpha1.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			Color:       category.Color,
		},
	}, nil
}

func (s *Server) ListCategories(ctx context.Context, req *chorerewardsv1alpha1.ListCategoriesRequest) (*chorerewardsv1alpha1.ListCategoriesResponse, error) {
	categories, err := s.dbManager.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	c := make([]*chorerewardsv1alpha1.Category, len(categories))
	for i, category := range categories {
		c[i] = &chorerewardsv1alpha1.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			Color:       category.Color,
		}
	}

	return &chorerewardsv1alpha1.ListCategoriesResponse{
		Categories: c,
	}, nil
}

func (s *Server) CreateTask(ctx context.Context, req *chorerewardsv1alpha1.CreateTaskRequest) (*chorerewardsv1alpha1.CreateTaskResponse, error) {
	task, err := s.dbManager.CreateTask(ctx, db.Task{
		CategoryID:   req.GetTask().GetCategoryId(),
		AssigneeID:   req.GetTask().GetAssigneeId(),
		Name:         req.GetTask().GetName(),
		Description:  req.GetTask().GetDescription(),
		Points:       req.GetTask().GetPoints(),
		IsRepeatable: req.GetTask().GetIsRepeatable(),
	})
	if err != nil {
		return nil, err
	}

	return &chorerewardsv1alpha1.CreateTaskResponse{
		Task: &chorerewardsv1alpha1.Task{
			Id:           task.ID,
			CategoryId:   task.CategoryID,
			AssigneeId:   task.AssigneeID,
			Name:         task.Name,
			Description:  task.Description,
			Points:       task.Points,
			IsRepeatable: task.IsRepeatable,
		},
	}, nil
}

func (s *Server) ListTasks(ctx context.Context, req *chorerewardsv1alpha1.ListTasksRequest) (*chorerewardsv1alpha1.ListTasksResponse, error) {
	tasks, err := s.dbManager.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	t := make([]*chorerewardsv1alpha1.Task, len(tasks))
	for i, task := range tasks {
		t[i] = &chorerewardsv1alpha1.Task{
			Id:           task.ID,
			CategoryId:   task.CategoryID,
			AssigneeId:   task.AssigneeID,
			Name:         task.Name,
			Description:  task.Description,
			Points:       task.Points,
			IsRepeatable: task.IsRepeatable,
		}
	}

	return &chorerewardsv1alpha1.ListTasksResponse{
		Tasks: t,
	}, nil
}

func (s *Server) AddTaskToFeed(ctx context.Context, req *chorerewardsv1alpha1.AddTaskToFeedRequest) (*chorerewardsv1alpha1.AddTaskToFeedResponse, error) {
	taskFeed, err := s.dbManager.CreateTaskFeed(ctx, db.TaskFeed{
		AssigneeID: req.GetTaskFeed().GetAssigneeId(),
		TaskID:     req.GetTaskFeed().GetTaskId(),
		IsComplete: req.GetTaskFeed().GetIsComplete(),
		IsApproved: req.GetTaskFeed().GetIsApproved(),
		Points:     req.GetTaskFeed().GetPoints(),
	})
	if err != nil {
		return nil, err
	}

	return &chorerewardsv1alpha1.AddTaskToFeedResponse{
		TaskFeed: &chorerewardsv1alpha1.TaskFeed{
			Id:         taskFeed.ID,
			TaskId:     taskFeed.TaskID,
			IsComplete: taskFeed.IsComplete,
			IsApproved: taskFeed.IsApproved,
			Points:     taskFeed.Points,
			AssigneeId: taskFeed.AssigneeID,
		},
	}, nil
}

func (s *Server) ListTasksFeed(ctx context.Context, req *chorerewardsv1alpha1.ListTasksFeedRequest) (*chorerewardsv1alpha1.ListTasksFeedResponse, error) {
	tasksFeed, err := s.dbManager.ListTasksFeed(ctx)
	if err != nil {
		return nil, err
	}

	tf := make([]*chorerewardsv1alpha1.TaskFeed, len(tasksFeed))
	for i, tfeed := range tasksFeed {
		tf[i] = &chorerewardsv1alpha1.TaskFeed{
			Id:         tfeed.ID,
			TaskId:     tfeed.TaskID,
			IsComplete: tfeed.IsComplete,
			IsApproved: tfeed.IsApproved,
			Points:     tfeed.Points,
			AssigneeId: tfeed.AssigneeID,
		}
	}

	return &chorerewardsv1alpha1.ListTasksFeedResponse{
		TaskFeed: tf,
	}, nil
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

func (s *Server) ListUsers(ctx context.Context, req *chorerewardsv1alpha1.ListUsersRequest) (*chorerewardsv1alpha1.ListUsersResponse, error) {
	users, err := s.dbManager.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	u := make([]*chorerewardsv1alpha1.User, len(users))
	for i, usr := range users {
		u[i] = &chorerewardsv1alpha1.User{
			Id:       usr.ID,
			Username: usr.Username,
			Email:    usr.Email,
			IsAdmin:  usr.IsAdmin,
			IsParent: usr.IsParent,
			Avatar:   usr.Avatar,
			Points:   usr.Points,
			IsActive: usr.IsActive,
		}
	}

	return &chorerewardsv1alpha1.ListUsersResponse{
		Users: u,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *chorerewardsv1alpha1.LoginRequest) (*chorerewardsv1alpha1.LoginResponse, error) {
	if req.GetUsername() == "" {
		return nil, status.Error(codes.Internal, "Username cannot be empty")
	}

	if req.GetPin() != 0 && req.GetPassword() != "" {
		return nil, status.Error(codes.Internal, "Specify either Pin OR Password, not both")
	}

	if req.GetPin() == 0 && req.GetPassword() == "" {
		return nil, status.Error(codes.Internal, "Specify either Pin OR Password")
	}

	user, err := s.dbManager.GetUser(ctx, req.GetUsername())
	if err != nil {
		// errors.As is the equivalent of a type assertion
		// if e, ok := err.(*errNotFound); ok
		if errors.As(err, &errNotFound) {
			return nil, status.Error(codes.Internal, "incorrect username or password")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	var authenticated bool

	if req.GetPin() != 0 {
		authenticated = auth.PasswordMatches([]byte(user.Pin), []byte(string(req.GetPin())))
	} else {
		authenticated = auth.PasswordMatches([]byte(user.Password), []byte(req.GetPassword()))
	}

	if !authenticated {
		return nil, status.Error(codes.PermissionDenied, "incorrect username or password")
	}

	token, err := s.tokenManager.CreateToken(req.GetUsername())
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create Token")
	}

	return &chorerewardsv1alpha1.LoginResponse{
		Token:    token,
		IsAdmin:  user.IsAdmin,
		IsParent: user.IsParent,
	}, nil
}
