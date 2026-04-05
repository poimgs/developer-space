package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/developer-space/api/internal/model"
)

type MessageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

func (r *MessageRepository) Create(ctx context.Context, channelID, memberID uuid.UUID, content string) (*model.MessageWithAuthor, error) {
	var m model.MessageWithAuthor
	err := r.pool.QueryRow(ctx,
		`WITH inserted AS (
			INSERT INTO messages (channel_id, member_id, content)
			VALUES ($1, $2, $3)
			RETURNING id, channel_id, member_id, content, created_at, updated_at
		)
		SELECT i.id, i.channel_id, i.member_id, i.content, i.created_at, i.updated_at, m.name
		FROM inserted i
		JOIN members m ON m.id = i.member_id`,
		channelID, memberID, content,
	).Scan(&m.ID, &m.ChannelID, &m.MemberID, &m.Content, &m.CreatedAt, &m.UpdatedAt, &m.AuthorName)
	if err != nil {
		return nil, fmt.Errorf("inserting message: %w", err)
	}
	return &m, nil
}

func (r *MessageRepository) ListByChannel(ctx context.Context, channelID uuid.UUID, cursor *time.Time, limit int) (*model.MessagePage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var args []any
	var query string

	if cursor != nil {
		query = `SELECT msg.id, msg.channel_id, msg.member_id, msg.content, msg.created_at, msg.updated_at, m.name
			FROM messages msg
			JOIN members m ON m.id = msg.member_id
			WHERE msg.channel_id = $1 AND msg.created_at < $2
			ORDER BY msg.created_at DESC
			LIMIT $3`
		args = []any{channelID, *cursor, limit + 1}
	} else {
		query = `SELECT msg.id, msg.channel_id, msg.member_id, msg.content, msg.created_at, msg.updated_at, m.name
			FROM messages msg
			JOIN members m ON m.id = msg.member_id
			WHERE msg.channel_id = $1
			ORDER BY msg.created_at DESC
			LIMIT $2`
		args = []any{channelID, limit + 1}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing messages: %w", err)
	}
	defer rows.Close()

	var messages []model.MessageWithAuthor
	for rows.Next() {
		var m model.MessageWithAuthor
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.MemberID, &m.Content, &m.CreatedAt, &m.UpdatedAt, &m.AuthorName); err != nil {
			return nil, fmt.Errorf("scanning message: %w", err)
		}
		messages = append(messages, m)
	}

	page := &model.MessagePage{}
	if len(messages) > limit {
		messages = messages[:limit]
		cursorStr := messages[limit-1].CreatedAt.Format(time.RFC3339Nano)
		page.Cursor = &cursorStr
	}

	if messages == nil {
		messages = []model.MessageWithAuthor{}
	}
	page.Messages = messages
	return page, nil
}
