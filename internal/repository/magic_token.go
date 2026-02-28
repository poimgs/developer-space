package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/developer-space/api/internal/model"
)

type MagicTokenRepository struct {
	pool *pgxpool.Pool
}

func NewMagicTokenRepository(pool *pgxpool.Pool) *MagicTokenRepository {
	return &MagicTokenRepository{pool: pool}
}

func (r *MagicTokenRepository) Create(ctx context.Context, memberID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.MagicToken, error) {
	var t model.MagicToken
	err := r.pool.QueryRow(ctx,
		`INSERT INTO magic_tokens (member_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, member_id, token_hash, expires_at, used_at, created_at`,
		memberID, tokenHash, expiresAt,
	).Scan(&t.ID, &t.MemberID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting magic token: %w", err)
	}
	return &t, nil
}

func (r *MagicTokenRepository) FindValidByHash(ctx context.Context, tokenHash string) (*model.MagicToken, error) {
	var t model.MagicToken
	err := r.pool.QueryRow(ctx,
		`SELECT id, member_id, token_hash, expires_at, used_at, created_at
		 FROM magic_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()`,
		tokenHash,
	).Scan(&t.ID, &t.MemberID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("finding magic token: %w", err)
	}
	return &t, nil
}

func (r *MagicTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE magic_tokens SET used_at = now() WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("marking token used: %w", err)
	}
	return nil
}

func (r *MagicTokenRepository) CountRecentByEmail(ctx context.Context, email string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM magic_tokens mt
		 JOIN members m ON mt.member_id = m.id
		 WHERE m.email = $1 AND mt.created_at > now() - interval '1 hour'`,
		email,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting recent tokens: %w", err)
	}
	return count, nil
}

func (r *MagicTokenRepository) CleanExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM magic_tokens WHERE expires_at < now() OR used_at IS NOT NULL`,
	)
	if err != nil {
		return 0, fmt.Errorf("cleaning expired tokens: %w", err)
	}
	return tag.RowsAffected(), nil
}
