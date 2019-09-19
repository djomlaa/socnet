package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	reEmail    = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	reUsername = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,17}$`)
	// ErrUserNotFound used when the user not found on the db.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidEmail used when the email is not valid.
	ErrInvalidEmail = errors.New("invalid email")
	// ErrInvalidUsername used when the username is not valid.
	ErrInvalidUsername = errors.New("invalid username")
	// ErrEmailTaken used when email already exists.
	ErrEmailTaken = errors.New("email is taken")
	// ErrUsernameTaken used when username already exists.
	ErrUsernameTaken = errors.New("username is taken")
	// ErrForbiddenFollow used when you try to follow yourself.
	ErrForbiddenFollow = errors.New("cannot follow yourself")
)

// User model
type User struct {
	ID       int64  `json:"id,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

// UserProfile model
type UserProfile struct {
	User
	Email          string `json:"email,omitempty"`
	FollowersCount int    `json:"followers_count"`
	FolloweesCount int    `json:"followees_count"`
	Me             bool   `json:"me,omitempty"`
	Following      bool   `json:"following,omitempty"`
	Followeed      bool   `json:"followeed,omitempty"`
}

// ToggleFollowOutput response
type ToggleFollowOutput struct {
	Following      bool `json:"following"`
	FollowersCount int  `json:"followers_count"`
}

// CreateUser inserts a user into db
func (s *Service) CreateUser(ctx context.Context, email, username string) error {

	email = strings.TrimSpace(email)

	if !reEmail.MatchString(email) {
		return ErrInvalidEmail
	}

	username = strings.TrimSpace(username)
	if !reUsername.MatchString(username) {
		return ErrInvalidUsername
	}

	query := "INSERT INTO users (email, username) VALUES ($1, $2)"
	_, err := s.db.ExecContext(ctx, query, email, username)

	unique := isUniqueViolation(err)

	if unique && strings.Contains((err.Error()), "email") {
		return ErrEmailTaken
	}

	if unique && strings.Contains((err.Error()), "username") {
		return ErrUsernameTaken
	}

	if err != nil {
		return fmt.Errorf("could not insert user: %v", err)
	}

	return nil
}

// User profile
func (s *Service) User(ctx context.Context, username string) (UserProfile, error) {

	var u UserProfile

	username = strings.TrimSpace(username)
	if !reUsername.MatchString(username) {
		return u, ErrInvalidUsername
	}

	uid, auth := ctx.Value(KeyAuthUserID).(int64)

	args := []interface{}{username}
	dest := []interface{}{&u.ID, &u.Email, &u.FollowersCount, &u.FolloweesCount}

	query := "SELECT id, email, followers_count, followees_count "
	if auth {
		query += ", " +
			"followers.follower_id IS NOT NULL as following, " +
			"followees.followee_id IS NOT NULL as followeed "
		dest = append(dest, &u.Following, &u.Followeed)
	}

	query += "FROM users "
	if auth {
		query += "LEFT JOIN follows AS followers on followers.follower_id = $2 AND followers.followee_id = users.id " +
			"LEFT JOIN follows AS followees on followees.follower_id = users.id AND followees.followee_id = $2 "
		args = append(args, uid)
	}
	query += "WHERE username =$1"

	err := s.db.QueryRowContext(ctx, query, args...).Scan(dest...)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query select user %v", err)
	}

	u.Username = username
	u.Me = auth && uid == u.ID
	if !u.Me {
		u.ID = 0
		u.Email = ""
	}

	return u, nil
}

// ToggleFollow between two users
func (s *Service) ToggleFollow(ctx context.Context, username string) (ToggleFollowOutput, error) {
	var out ToggleFollowOutput

	followerID, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return out, ErrUnauthenticated
	}

	username = strings.TrimSpace(username)
	if !reUsername.MatchString(username) {
		return out, ErrInvalidUsername
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, fmt.Errorf("could not begin tx: %v", err)
	}

	defer tx.Rollback()

	var followeeID int64
	query := "SELECT id FROM users WHERE username = $1"
	err = tx.QueryRowContext(ctx, query, username).Scan(&followeeID)
	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err != nil {
		return out, fmt.Errorf("could not query select user id from followee username %v", err)
	}

	if followeeID == followerID {
		return out, ErrForbiddenFollow
	}

	query = "SELECT EXISTS (SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2)"
	err = tx.QueryRowContext(ctx, query, followerID, followeeID).Scan(&out.Following)
	if err != nil {
		return out, fmt.Errorf("could not query select existence of follow %v", err)
	}

	if out.Following {
		query = "DELETE FROM follows WHERE follower_id =$1 AND followee_id =$2"
		if _, err = tx.ExecContext(ctx, query, followerID, followeeID); err != nil {
			return out, fmt.Errorf("could not delete follow: %v", err)
		}

		query = "UPDATE users SET followees_count = followees_count - 1 WHERE id = $1"
		if _, err = tx.ExecContext(ctx, query, followerID); err != nil {
			return out, fmt.Errorf("could not update follower followees count (-): %v", err)
		}

		query = "UPDATE users SET followers_count = followers_count - 1 WHERE id = $1 RETURNING followers_count"
		if err = tx.QueryRowContext(ctx, query, followeeID).Scan(&out.FollowersCount); err != nil {
			return out, fmt.Errorf("could not update followee followers count (-): %v", err)
		}
	} else {
		query = "INSERT INTO follows (follower_id, followee_id) VALUES ($1, $2)"
		if _, err = tx.ExecContext(ctx, query, followerID, followeeID); err != nil {
			return out, fmt.Errorf("could not insert follow: %v", err)
		}

		query = "UPDATE users SET followees_count = followees_count  + 1 WHERE id = $1"
		if _, err = tx.ExecContext(ctx, query, followerID); err != nil {
			return out, fmt.Errorf("could not update follower followees count (+): %v", err)
		}

		query = "UPDATE users SET followers_count = followers_count + 1 WHERE id = $1 RETURNING followers_count"
		if err = tx.QueryRowContext(ctx, query, followeeID).Scan(&out.FollowersCount); err != nil {
			return out, fmt.Errorf("could not update followee followers count (+) %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return out, fmt.Errorf("could not commit toggle follow: %v", err)
	}

	out.Following = !out.Following

	if out.Following {
		// TODO: notify followee
	}

	return out, nil
}
