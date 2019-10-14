package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// TokenLifespan until tokens are valid
	TokenLifespan = time.Hour * 24 * 14
	// KeyAuthUserID to use in context
	KeyAuthUserID key = "auth_user_id"
)

var (
	// ErrUnauthenticated used when there is no autheniticated user in context
	ErrUnauthenticated = errors.New("unauthenticated")
)

type key string

//LoginOutput response
type LoginOutput struct {
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	AuthUser  User      `json:"auth_user,omitempty"`
}

// AuthUserID from Token
func (s *Service) AuthUserID(token string) (int64, error) {

	str, err := s.codec.DecodeToString(token)
	if err != nil {
		return 0, fmt.Errorf("could not decode token: %v", err)
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse auth user id from token: %v", err)
	}
	return i, nil
}

//Login insecurely
func (s *Service) Login(ctx context.Context, email string) (LoginOutput, error) {

	var out LoginOutput

	email = strings.TrimSpace(email)
	if !reEmail.MatchString(email) {
		return out, ErrInvalidEmail
	}

	var avatar sql.NullString
	query := "SELECT id, username, avatar FROM users WHERE email = $1"
	err := s.db.QueryRowContext(ctx, query, email).Scan(&out.AuthUser.ID, &out.AuthUser.Username, &avatar)

	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err != nil {
		return out, fmt.Errorf("could not query select user: %v", err)
	}

	if avatar.Valid {
		avatarURL := s.origin + "/img/avatars/" + avatar.String
		out.AuthUser.AvatarURL = &avatarURL
	}

	out.Token, err = s.codec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))

	if err != nil {
		return out, fmt.Errorf("could not create token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifespan)

	return out, nil
}

// AuthUser from context
func (s *Service) AuthUser(ctx context.Context) (User, error) {
	var u User
	uid, ok := ctx.Value(KeyAuthUserID).(int64)

	if !ok {
		return u, ErrUnauthenticated
	}

	return s.userByID(ctx, uid)
}
