package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"newFeatures/models"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (a *AuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO users (name, email, phone, password, role) VALUES ($1, $2, $3, $4, $5)`, user.Name, user.Email, user.Phone, user.Password, user.Role)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthRepository) UserByPhone(ctx context.Context, user *models.User) (*models.User, error) {
	var userDB models.User
	result := a.db.QueryRowContext(ctx,
		`SELECT id, name, email, phone, password, role FROM users WHERE phone = $1`, user.Phone)
	if err := result.Scan(&userDB.Id, &userDB.Name, &userDB.Email, &userDB.Phone, &userDB.Password, &userDB.Role); err != nil {
		return nil, err
	}
	return &userDB, nil
}

func (a *AuthRepository) UserById(ctx context.Context, userID int) (*models.ResponseUser, error) {
	var user models.ResponseUser
	result := a.db.QueryRowContext(ctx, `SELECT id, name, email, phone FROM users WHERE id = $1`, userID)
	if err := result.Scan(&user.Id, &user.Name, &user.Email, &user.Phone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}
	return &user, nil
}

func (a *AuthRepository) Users(ctx context.Context, page, limit int64) ([]models.ResponseUser, error) {
	offset := (page - 1) * limit
	query := fmt.Sprintf("SELECT id, name, email, phone FROM users OFFSET %d LIMIT %d", offset, limit)

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.ResponseUser
	for rows.Next() {
		var user models.ResponseUser
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Phone)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (a *AuthRepository) UpdateUser(ctx context.Context, inputUser *models.ResponseUser) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE users SET name = $1, email = $2, phone = $3 WHERE id = $4", inputUser.Name, inputUser.Email, inputUser.Phone, inputUser.Id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (a *AuthRepository) DeleteUser(ctx context.Context, userID int) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, userID)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthRepository) UserRoleById(userID int) (*models.User, error) {
	var user models.User
	user.Id = userID
	err := a.db.QueryRow(`SELECT role FROM users WHERE id = $1`, user.Id).Scan(&user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, err
	}
	return &user, nil
}

func (a *AuthRepository) RestorePassword(ctx context.Context, restore *models.RestorePassword) error {
	_, err := a.db.ExecContext(ctx, "UPDATE users SET password = $1 WHERE email = $2", restore.Password, restore.Email)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthRepository) CheckByEmail(ctx context.Context, restore *models.RestorePassword) error {
	var exist bool
	query := "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)"
	row := a.db.QueryRowContext(ctx, query, restore.Email)
	if err := row.Scan(&exist); err != nil {
		log.Printf("Error while scanning for email: %s", err)
		return fmt.Errorf("error while scanning for email")
	}
	if !exist {
		return fmt.Errorf("user with this email does not exist")
	}
	return nil
}
