package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrUserNotFound is returned when a user lookup yields no result.
var ErrUserNotFound = errors.New("user not found")

// ErrEmailTaken is returned when registering with an already-used email.
var ErrEmailTaken = errors.New("email already in use")

// Repository defines the data access contract for the auth domain.
type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	IncrementLoginFailed(ctx context.Context, userID uuid.UUID) error
	ResetLoginFailed(ctx context.Context, userID uuid.UUID) error
	LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error

	CreateSession(ctx context.Context, session *UserSession) error
	FindSessionByJTI(ctx context.Context, jti string) (*UserSession, error)
	FindSessionByRefreshJTI(ctx context.Context, refreshJTI string) (*UserSession, error)
	RevokeSessionByJTI(ctx context.Context, jti string) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error

	CreateResetToken(ctx context.Context, token *PasswordResetToken) error
	FindResetToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error
}

// postgresRepository is the GORM/PostgreSQL implementation of Repository.
type postgresRepository struct {
	db *gorm.DB
}

// NewRepository creates a new auth Repository backed by PostgreSQL.
func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *postgresRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *postgresRepository) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *postgresRepository) UpdateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *postgresRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"login_failed":  0,
			"locked_until":  nil,
		}).Error
}

func (r *postgresRepository) IncrementLoginFailed(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ?", userID).
		UpdateColumn("login_failed", gorm.Expr("login_failed + 1")).Error
}

func (r *postgresRepository) ResetLoginFailed(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{"login_failed": 0, "locked_until": nil}).Error
}

func (r *postgresRepository) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ?", userID).
		Update("locked_until", until).Error
}

func (r *postgresRepository) CreateSession(ctx context.Context, session *UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *postgresRepository) FindSessionByJTI(ctx context.Context, jti string) (*UserSession, error) {
	var session UserSession
	err := r.db.WithContext(ctx).Where("jti = ?", jti).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("session not found")
	}
	return &session, err
}

func (r *postgresRepository) FindSessionByRefreshJTI(ctx context.Context, refreshJTI string) (*UserSession, error) {
	var session UserSession
	err := r.db.WithContext(ctx).Where("refresh_jti = ?", refreshJTI).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("session not found")
	}
	return &session, err
}

func (r *postgresRepository) RevokeSessionByJTI(ctx context.Context, jti string) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("jti = ?", jti).
		Update("is_revoked", true).Error
}

func (r *postgresRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Update("is_revoked", true).Error
}

func (r *postgresRepository) DeleteExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&UserSession{}).Error
}

func (r *postgresRepository) CreateResetToken(ctx context.Context, token *PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *postgresRepository) FindResetToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	var token PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("token not found or expired")
	}
	return &token, err
}

func (r *postgresRepository) MarkResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&PasswordResetToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now).Error
}
