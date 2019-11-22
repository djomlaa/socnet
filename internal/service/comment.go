package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Comment model
type Comment struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"-"`
	PostID     int64     `json:"-"`
	Content    string    `json:"content"`
	LikesCount int       `json:"likes_count"`
	CreatedAt  time.Time `json:"createdAt"`
	User       *User     `json:"user,omitempty"`
	Mine       bool      `json:"mine"`
	Liked      bool      `json:"liked"`
}

var (
	// ErrCommentNotFound denotes a post that was not found
	ErrCommentNotFound = errors.New("comment not found")
)

// CreateComment on post
func (s *Service) CreateComment(ctx context.Context, postID int64, content string) (Comment, error) {
	var c Comment
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return c, ErrUnauthenticated
	}

	content = strings.TrimSpace(content)
	if content == "" || len([]rune(content)) == 480 {
		return c, ErrInvalidContent
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return c, fmt.Errorf("could not begin tx: %v", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO comments (user_id, post_id, content) VALUES ($1, $2, $3)
			  RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, query, uid, postID, content).Scan(&c.ID, &c.CreatedAt)
	if isForeignKeyViolation(err) {
		return c, ErrPostNotFound
	}
	if err != nil {
		return c, fmt.Errorf("could not insert comment: %v", err)
	}

	c.UserID = uid
	c.PostID = postID
	c.Content = content
	c.Mine = true

	query = "UPDATE posts SET comments_count = comments_count + 1 WHERE id =$1"
	if _, err = tx.ExecContext(ctx, query, postID); err != nil {
		return c, fmt.Errorf("could not update and increase comments count comment: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return c, fmt.Errorf("could not commit tx: %v", err)
	}

	return c, nil
}

// Comments from a post in descending order with backward pagination
func (s *Service) Comments(ctx context.Context, postID int64, last int, before int64) ([]Comment, error) {
	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	last = normalizePageSize(last)
	query, args, err := buildQuery(`
		SELECT c.id, c.content, c.likes_count, c.created_at, u.username, u.avatar
		{{if .auth}}
		, c.user_id =@uid as mine
		, cl.user_id IS NOT NULL AS likes
		{{end}}
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		{{if .auth}}
		LEFT JOIN comment_likes cl ON cl.comment_id = c.id AND cl.user_id =@uid
		{{end}}
		WHERE c.post_id = @post_id
		{{if .before}}AND c.id < @before{{end}}
		ORDER BY c.created_at DESC
		LIMIT @last`,
		map[string]interface{}{
			"auth":    auth,
			"uid":     uid,
			"post_id": postID,
			"before":  before,
			"last":    last,
		})

	if err != nil {
		return nil, fmt.Errorf("could not build comments sql query: %v", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query select comments: %v", err)
	}

	defer rows.Close()

	cc := make([]Comment, 0, last)
	for rows.Next() {
		var c Comment
		var u User
		var avatar sql.NullString
		dest := []interface{}{&c.ID, &c.Content, &c.LikesCount, &c.CreatedAt, &u.Username, &avatar}
		if auth {
			dest = append(dest, &c.Mine, &c.Liked)
		}
		if err = rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("could not scan comments: %v", err)
		}

		if avatar.Valid {
			avatarURL := s.origin + "/img/avatars/" + avatar.String
			u.AvatarURL = &avatarURL
		}
		c.User = &u
		cc = append(cc, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate comment rows: , %v", err)
	}

	return cc, nil
}

// ToggleCommentLike -
func (s *Service) ToggleCommentLike(ctx context.Context, commentID int64) (ToggleLikeOutput, error) {
	var out ToggleLikeOutput
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return out, ErrUnauthenticated
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, fmt.Errorf("could not begin tx: %v", err)
	}

	defer tx.Rollback()

	query := `
		SELECT EXISTS (
			SELECT 1 FROM comment_likes WHERE user_id = $1 AND comment_id = $2
	)`

	if err := tx.QueryRowContext(ctx, query, uid, commentID).Scan(&out.Liked); err != nil {
		return out, fmt.Errorf("could not query select existence: %v", err)
	}

	if out.Liked {
		query = "DELETE FROM comment_likes WHERE user_id = $1 AND comment_id = $2"
		if _, err = tx.ExecContext(ctx, query, uid, commentID); err != nil {
			return out, fmt.Errorf("could not delete comment like: %v", err)
		}
		query = "UPDATE comments SET likes_count = likes_count - 1 WHERE id = $1 RETURNING likes_count"
		if err = tx.QueryRowContext(ctx, query, commentID).Scan(&out.LikesCount); err != nil {
			return out, fmt.Errorf("could not update and decerement comment likes count: %v", err)
		}
	} else {
		query = "INSERT into comment_likes (user_id, comment_id) VALUES ($1, $2)"
		_, err = tx.ExecContext(ctx, query, uid, commentID)

		if isForeignKeyViolation(err) {
			return out, ErrCommentNotFound
		}
		if err != nil {
			return out, fmt.Errorf("could not insert comment like: %v", err)
		}

		query = "UPDATE comments SET likes_count = likes_count + 1 WHERE id = $1 RETURNING likes_count"
		if err = tx.QueryRowContext(ctx, query, commentID).Scan(&out.LikesCount); err != nil {
			return out, fmt.Errorf("could not update and incerement comment likes count: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return out, fmt.Errorf("could not commit to toggle comment like: %v", err)
	}

	out.Liked = !out.Liked

	return out, nil
}
