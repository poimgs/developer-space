package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/developer-space/api/internal/model"
)

type ChannelRepository struct {
	pool *pgxpool.Pool
}

func NewChannelRepository(pool *pgxpool.Pool) *ChannelRepository {
	return &ChannelRepository{pool: pool}
}

func (r *ChannelRepository) Create(ctx context.Context, name, channelType string, sessionID *uuid.UUID, createdBy uuid.UUID) (*model.Channel, error) {
	var c model.Channel
	err := r.pool.QueryRow(ctx,
		`INSERT INTO channels (name, type, session_id, created_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, type, session_id, created_by, created_at, updated_at`,
		name, channelType, sessionID, createdBy,
	).Scan(&c.ID, &c.Name, &c.Type, &c.SessionID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting channel: %w", err)
	}
	return &c, nil
}

func (r *ChannelRepository) List(ctx context.Context) ([]model.Channel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, type, session_id, created_by, created_at, updated_at
		 FROM channels
		 ORDER BY type ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("listing channels: %w", err)
	}
	defer rows.Close()

	var channels []model.Channel
	for rows.Next() {
		var c model.Channel
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.SessionID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning channel: %w", err)
		}
		channels = append(channels, c)
	}
	if channels == nil {
		channels = []model.Channel{}
	}
	return channels, nil
}

func (r *ChannelRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	var c model.Channel
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, type, session_id, created_by, created_at, updated_at
		 FROM channels WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Type, &c.SessionID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting channel by id: %w", err)
	}
	return &c, nil
}

func (r *ChannelRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*model.Channel, error) {
	var c model.Channel
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, type, session_id, created_by, created_at, updated_at
		 FROM channels WHERE session_id = $1`, sessionID,
	).Scan(&c.ID, &c.Name, &c.Type, &c.SessionID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting channel by session id: %w", err)
	}
	return &c, nil
}

func (r *ChannelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM channels WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting channel: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}
