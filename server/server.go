package server

import (
	"context"

	chorerewardsv1alpha1 "github.com/chorerewards/api/chorerewards/v1alpha1"
)

// Server is the implementation of the chorerewardsv1alpha1.ChoreRewardsServiceServer
type Server struct{}

func New() *Server {
	return &Server{}
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

func (s *Server) CreateUser(context.Context, *chorerewardsv1alpha1.CreateUserRequest) (*chorerewardsv1alpha1.CreateUserResponse, error) {
	return nil, nil
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
