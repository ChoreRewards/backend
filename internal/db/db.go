package db

import (
	"context"
	"fmt"

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
		return u, errors.Wrap(err, "unable to add user")
	}

	return u, nil
}

func (d *Manager) ListUsers(ctx context.Context) ([]User, error) {
	users := make([]User, 0)

	rows, err := d.pool.Query(ctx, "Select id, username, email, is_admin, is_parent, avatar, points, is_active FROM users")
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
