package service

import (
	"context"
	"database/sql"
	"fmt"
)

// TimelineItem model
type TimelineItem struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"-"`
	PostID int64 `json:"-"`
	Post   Post  `json:"post"`
}

// Timeline -
func (s *Service) Timeline(ctx context.Context, last int, before int) ([]TimelineItem, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return nil, ErrUnauthenticated
	}
	last = normalizePageSize(last)
	query, args, err := buildQuery(`
		SELECT t.id, p.id, p.content, p.spoiler_of, p.nsfw, p.likes_count, p.comments_count, p.created_at
		, p.user_id = @uid AS mine
		, pl.user_id IS NOT NULL AS liked
		, u.username, u.avatar
		FROM timeline t
		INNER JOIN posts p ON t.post_id = p.id
		INNER JOIN users u ON p.user_id = u.id
		LEFT JOIN post_likes pl on pl.user_id = p.user_id and pl.post_id = p.id
		WHERE t.user_id = @uid
		{{if .before}}	AND t.id < @before {{end}}
		ORDER BY created_at DESC
		LIMIT @last
	`, map[string]interface{}{
		"uid":    uid,
		"last":   last,
		"before": before,
	})
	if err != nil {
		return nil, fmt.Errorf("could not build timeline sql query: %v", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query select timeline: %v", err)
	}
	defer rows.Close()

	var u User
	var avatar sql.NullString
	tt := make([]TimelineItem, 0, last)
	for rows.Next() {
		var ti TimelineItem
		dest := []interface{}{
			&ti.ID,
			&ti.Post.ID,
			&ti.Post.Content,
			&ti.Post.SpoilerOf,
			&ti.Post.NSFW,
			&ti.Post.LikesCount,
			&ti.Post.CommentsCount,
			&ti.Post.CreatedAt,
			&ti.Post.Mine,
			&ti.Post.Liked,
			&u.Username,
			&avatar,
		}

		if err = rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("could not scan timeline item: %v", err)
		}

		if avatar.Valid {
			avatarURL := s.origin + "/img/avatars/" + avatar.String
			u.AvatarURL = &avatarURL
		}

		ti.Post.User = &u
		tt = append(tt, ti)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate timeline rows: %v", err)
	}

	return tt, nil
}
