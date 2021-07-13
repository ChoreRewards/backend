package db

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type Manager struct {
	pool *pgxpool.Pool
}

type Category struct {
	ID          int32
	Color       string
	Name        string
	Description string
}

type Task struct {
	ID           int32
	CategoryID   int32
	AssigneeID   int32
	Name         string
	Description  string
	Points       int32
	IsRepeatable bool
}

type TaskFeed struct {
	ID          int32
	AssigneeID  int32
	TaskID      int32
	IsComplete  bool
	IsApproved  bool
	CompletedAt timestamp.Timestamp
	Points      int32
}

type User struct {
	ID       int32
	Username string
	Email    string
	IsAdmin  bool
	IsParent bool
	Avatar   string
	Points   int32
	Password string
	Pin      string
	IsActive bool
}

var _ error = (*ErrNotFound)(nil) // ensure CustomError implements error

type ErrNotFound struct {
	message string
}

func (c *ErrNotFound) Error() string {
	return c.message
}

func New(c Config) (*Manager, error) {
	if c.Host == "" {
		return nil, errors.New("host not defined")
	}

	if c.Port == 0 {
		return nil, errors.New("port not defined")
	}

	if c.Username == "" {
		return nil, errors.New("user not defined")
	}

	if c.Password == "" {
		return nil, errors.New("password not defined")
	}

	if c.Database == "" {
		return nil, errors.New("database not defined")
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Database)

	pool, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	return &Manager{pool: pool}, nil
}

func (d *Manager) CreateCategory(ctx context.Context, category Category) (Category, error) {
	c := Category{}

	err := d.pool.QueryRow(
		ctx,
		"INSERT INTO categories(color, name, description) VALUES($1, $2, $3) RETURNING id, color, name, description",
		category.Color, category.Name, category.Description,
	).Scan(&c.ID, &c.Color, &c.Name, &c.Description)
	if err != nil {
		return c, errors.Wrap(err, "unable to add category")
	}

	logrus.WithFields(logrus.Fields{
		"id": c.ID,
	}).Info("Category inserted successfully")

	return c, nil
}

func (d *Manager) GetCategory(ctx context.Context, name string) (Category, error) {
	c := Category{}

	err := d.pool.QueryRow(ctx, "SELECT id, color, name, description FROM categories WHERE name=$1", name).
		Scan(&c.ID, &c.Color, &c.Name, &c.Description)
	if err != nil {
		return c, errors.Wrap(err, "unable to get category")
	}

	return c, nil
}

func (d *Manager) ListCategories(ctx context.Context) ([]Category, error) {
	categories := make([]Category, 0)

	rows, err := d.pool.Query(ctx, "SELECT id, color, name, description FROM categories")
	if err != nil {
		return categories, errors.Wrap(err, "unable to get users")
	}

	rowCount := 0
	for rows.Next() {
		c := Category{}

		if err := rows.Scan(&c.ID, &c.Color, &c.Name, &c.Description); err != nil {
			return nil, errors.Wrap(err, "unable to scan row")
		}

		categories = append(categories, c)

		rowCount++
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "erroring reading rows")
	}

	logrus.WithFields(logrus.Fields{"rowCount": rowCount}).Info("Categories queried successfully")

	return categories, nil
}

func (d *Manager) CreateTask(ctx context.Context, task Task) (Task, error) {
	t := Task{}

	err := d.pool.QueryRow(
		ctx,
		"INSERT INTO tasks(category_id, assignee_id, name, description, points, is_repeatable) VALUES($1, $2, $3, $4, $5, $6) RETURNING id, category_id, assignee_id, name, description, points, is_repeatable",
		task.CategoryID, task.AssigneeID, task.Name, task.Description, task.Points, task.IsRepeatable,
	).Scan(&t.ID, &t.CategoryID, &t.AssigneeID, &t.Name, &t.Description, &t.Points, &t.IsRepeatable)
	if err != nil {
		return t, errors.Wrap(err, "unable to add task")
	}

	logrus.WithFields(logrus.Fields{
		"id": t.ID,
	}).Info("Task inserted successfully")

	return t, nil
}

func (d *Manager) GetTask(ctx context.Context, name string) (Task, error) {
	t := Task{}

	err := d.pool.QueryRow(ctx, "SELECT category_id, assignee_id, name, description, points, is_repeatable FROM tasks WHERE name=$1", name).
		Scan(&t.ID, &t.CategoryID, &t.AssigneeID, &t.Name, &t.Description, &t.Points, &t.IsRepeatable)
	if err != nil {
		return t, errors.Wrap(err, "unable to get task")
	}

	return t, nil
}

func (d *Manager) ListTasks(ctx context.Context) ([]Task, error) {
	tasks := make([]Task, 0)

	rows, err := d.pool.Query(ctx, "SELECT category_id, assignee_id, name, description, points, is_repeatable FROM tasks ")
	if err != nil {
		return tasks, errors.Wrap(err, "unable to get tasks")
	}

	rowCount := 0
	for rows.Next() {
		t := Task{}

		if err := rows.Scan(&t.ID, &t.CategoryID, &t.AssigneeID, &t.Name, &t.Description, &t.Points, &t.IsRepeatable); err != nil {
			return nil, errors.Wrap(err, "unable to scan row")
		}

		tasks = append(tasks, t)

		rowCount++
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "erroring reading rows")
	}

	logrus.WithFields(logrus.Fields{"rowCount": rowCount}).Info("Tasks queried successfully")

	return tasks, nil
}

/*

type TaskFeed struct {
	ID          int32
	AssigneeID  int32
	TaskID      int32
	IsComplete  bool
	IsApproved  bool
	CompletedAt timestamp.Timestamp
	Points      int32
}

*/

func (d *Manager) CreateTaskFeed(ctx context.Context, taskFeed TaskFeed) (TaskFeed, error) {
	tf := TaskFeed{}

	err := d.pool.QueryRow(
		ctx,
		"INSERT INTO tasks_feed(assignee_id, task_id, is_complete, is_approved, completed_at, points) VALUES($1, $2, $3, $4, $5, $6) RETURNING id, assignee_id, task_id, is_complete, is_approved, completed_at, points",
		taskFeed.AssigneeID, taskFeed.TaskID, taskFeed.IsComplete, taskFeed.IsApproved, taskFeed.CompletedAt, taskFeed.Points,
	).Scan(&tf.ID, &tf.AssigneeID, &tf.TaskID, &tf.IsComplete, &tf.IsApproved, &tf.CompletedAt, &tf.Points)
	if err != nil {
		return tf, errors.Wrap(err, "unable to add task feed")
	}

	logrus.WithFields(logrus.Fields{
		"id": tf.ID,
	}).Info("Task Feed inserted successfully")

	return tf, nil
}

func (d *Manager) ListTasksFeed(ctx context.Context) ([]TaskFeed, error) {
	tasksFeed := make([]TaskFeed, 0)

	rows, err := d.pool.Query(ctx, "SELECT id, assignee_id, task_id, is_complete, is_approved, completed_at, points FROM tasks_feed")
	if err != nil {
		return tasksFeed, errors.Wrap(err, "unable to get tasks feed")
	}

	rowCount := 0
	for rows.Next() {
		tf := TaskFeed{}

		if err := rows.Scan(&tf.ID, &tf.AssigneeID, &tf.TaskID, &tf.IsComplete, &tf.IsApproved, &tf.CompletedAt, &tf.Points); err != nil {
			return nil, errors.Wrap(err, "unable to scan row")
		}

		tasksFeed = append(tasksFeed, tf)

		rowCount++
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "erroring reading rows")
	}

	logrus.WithFields(logrus.Fields{"rowCount": rowCount}).Info("Tasks Feed queried successfully")

	return tasksFeed, nil
}

func (d *Manager) CreateUser(ctx context.Context, user User) (User, error) {
	u := User{}

	err := d.pool.QueryRow(
		ctx,
		"INSERT INTO users(username, email, is_admin, is_parent, avatar, password, pin, points, is_active) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, username, email, is_admin, is_parent, avatar, points, is_active",
		user.Username, user.Email, user.IsAdmin, user.IsParent, user.Avatar, user.Password, user.Pin, 0, true,
	).Scan(&u.ID, &u.Username, &u.Email, &u.IsAdmin, &u.IsParent, &u.Avatar, &u.Points, &u.IsActive)
	if err != nil {
		return u, errors.Wrap(err, "unable to add user")
	}

	logrus.WithFields(logrus.Fields{
		"id": u.ID,
	}).Info("User inserted successfully")

	return u, nil
}

func (d *Manager) GetUser(ctx context.Context, username string) (User, error) {
	u := User{}

	err := d.pool.QueryRow(ctx, "SELECT id, username, email, is_admin, is_parent, avatar, points, password, pin, is_active FROM users WHERE username=$1", username).
		Scan(&u.ID, &u.Username, &u.Email, &u.IsAdmin, &u.IsParent, &u.Avatar, &u.Points, &u.Password, &u.Pin, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return u, &ErrNotFound{message: "record not found"}
		}
		return u, errors.Wrap(err, "unable to get user")
	}

	return u, nil
}

func (d *Manager) ListUsers(ctx context.Context) ([]User, error) {
	users := make([]User, 0)

	rows, err := d.pool.Query(ctx, "SELECT id, username, email, is_admin, is_parent, avatar, points, is_active FROM users")
	if err != nil {
		return users, errors.Wrap(err, "unable to get users")
	}

	rowCount := 0
	for rows.Next() {
		u := User{}

		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.IsAdmin, &u.IsParent, &u.Avatar, &u.Points, &u.IsActive); err != nil {
			return nil, errors.Wrap(err, "unable to scan row")
		}

		users = append(users, u)

		rowCount++
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "erroring reading rows")
	}

	logrus.WithFields(logrus.Fields{"rowCount": rowCount}).Info("Users queried successfully")

	return users, nil
}
