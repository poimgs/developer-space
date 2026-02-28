package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/developer-space/api/internal/model"
)

type MemberRepository struct {
	pool *pgxpool.Pool
}

func NewMemberRepository(pool *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{pool: pool}
}

func (r *MemberRepository) Create(ctx context.Context, req model.CreateMemberRequest) (*model.Member, error) {
	var m model.Member
	err := r.pool.QueryRow(ctx,
		`INSERT INTO members (email, name, telegram_handle, is_admin)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, name, telegram_handle, is_admin, is_active, created_at, updated_at`,
		req.Email, req.Name, req.TelegramHandle, req.IsAdmin,
	).Scan(&m.ID, &m.Email, &m.Name, &m.TelegramHandle, &m.IsAdmin, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting member: %w", err)
	}
	return &m, nil
}

func (r *MemberRepository) List(ctx context.Context, activeFilter string) ([]model.Member, error) {
	var query string
	var args []any

	switch activeFilter {
	case "false":
		query = `SELECT id, email, name, telegram_handle, is_admin, is_active, created_at, updated_at
				 FROM members WHERE is_active = false ORDER BY name`
	case "all":
		query = `SELECT id, email, name, telegram_handle, is_admin, is_active, created_at, updated_at
				 FROM members ORDER BY name`
	default: // "true" or empty
		query = `SELECT id, email, name, telegram_handle, is_admin, is_active, created_at, updated_at
				 FROM members WHERE is_active = true ORDER BY name`
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	defer rows.Close()

	var members []model.Member
	for rows.Next() {
		var m model.Member
		if err := rows.Scan(&m.ID, &m.Email, &m.Name, &m.TelegramHandle, &m.IsAdmin, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, m)
	}

	if members == nil {
		members = []model.Member{}
	}

	return members, nil
}

func (r *MemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	var m model.Member
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, telegram_handle, is_admin, is_active, created_at, updated_at
		 FROM members WHERE id = $1`, id,
	).Scan(&m.ID, &m.Email, &m.Name, &m.TelegramHandle, &m.IsAdmin, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting member by id: %w", err)
	}
	return &m, nil
}

func (r *MemberRepository) GetByEmail(ctx context.Context, email string) (*model.Member, error) {
	var m model.Member
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, telegram_handle, is_admin, is_active, created_at, updated_at
		 FROM members WHERE email = $1`, email,
	).Scan(&m.ID, &m.Email, &m.Name, &m.TelegramHandle, &m.IsAdmin, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting member by email: %w", err)
	}
	return &m, nil
}

func (r *MemberRepository) Update(ctx context.Context, id uuid.UUID, req model.UpdateMemberRequest) (*model.Member, error) {
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.TelegramHandle != nil {
		setClauses = append(setClauses, fmt.Sprintf("telegram_handle = $%d", argIdx))
		args = append(args, *req.TelegramHandle)
		argIdx++
	}
	if req.IsAdmin != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_admin = $%d", argIdx))
		args = append(args, *req.IsAdmin)
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE members SET %s WHERE id = $%d
		 RETURNING id, email, name, telegram_handle, is_admin, is_active, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var m model.Member
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&m.ID, &m.Email, &m.Name, &m.TelegramHandle, &m.IsAdmin, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("updating member: %w", err)
	}
	return &m, nil
}

func (r *MemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM members WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *MemberRepository) HasRSVPs(ctx context.Context, memberID uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM rsvps WHERE member_id = $1`, memberID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking member rsvps: %w", err)
	}
	return count > 0, nil
}
