package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/developer-space/api/internal/model"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) (*model.SpaceSession, error) {
	var s model.SpaceSession
	err := r.pool.QueryRow(ctx,
		`INSERT INTO space_sessions (title, description, date, start_time, end_time, location, status, series_id, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, 'scheduled', $7, $8)
		 RETURNING id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), status, image_url, location, series_id, created_by, created_at, updated_at`,
		req.Title, req.Description, req.Date, req.StartTime, req.EndTime, req.Location, req.SeriesID, createdBy,
	).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting session: %w", err)
	}
	return &s, nil
}

func (r *SessionRepository) CreateBatch(ctx context.Context, sessions []model.CreateSessionRequest, createdBy uuid.UUID) ([]model.SpaceSession, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var result []model.SpaceSession
	for _, req := range sessions {
		var s model.SpaceSession
		err := tx.QueryRow(ctx,
			`INSERT INTO space_sessions (title, description, date, start_time, end_time, location, status, series_id, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, 'scheduled', $7, $8)
			 RETURNING id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), status, image_url, location, series_id, created_by, created_at, updated_at`,
			req.Title, req.Description, req.Date, req.StartTime, req.EndTime, req.Location, req.SeriesID, createdBy,
		).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("inserting batch session: %w", err)
		}
		result = append(result, s)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing batch: %w", err)
	}
	return result, nil
}

func (r *SessionRepository) List(ctx context.Context, from, to, status string, memberID *uuid.UUID) ([]model.SpaceSession, error) {
	args := []any{}
	conditions := []string{}
	argIdx := 1

	// Reserve $1 for memberID used in the EXISTS subquery
	var memberIDArg any
	if memberID != nil {
		memberIDArg = *memberID
	} else {
		memberIDArg = uuid.Nil
	}
	args = append(args, memberIDArg)
	argIdx++

	if from != "" {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argIdx))
		args = append(args, from)
		argIdx++
	} else {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argIdx))
		args = append(args, time.Now().Format("2006-01-02"))
		argIdx++
	}

	if to != "" {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argIdx))
		args = append(args, to)
		argIdx++
	} else {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argIdx))
		args = append(args, time.Now().AddDate(0, 0, 60).Format("2006-01-02"))
		argIdx++
	}

	switch status {
	case "scheduled":
		conditions = append(conditions, "s.status = 'scheduled'")
	case "shifted":
		conditions = append(conditions, "s.status = 'shifted'")
	case "canceled":
		conditions = append(conditions, "s.status = 'canceled'")
	case "all":
		// no status filter
	default:
		conditions = append(conditions, "s.status IN ('scheduled', 'shifted')")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(
		`SELECT s.id, s.title, s.description, s.date::text, to_char(s.start_time, 'HH24:MI'), to_char(s.end_time, 'HH24:MI'), s.status,
		        s.image_url, s.location, s.series_id, s.created_by, s.created_at, s.updated_at,
		        COALESCE(COUNT(r.id), 0) AS rsvp_count,
		        EXISTS(SELECT 1 FROM rsvps WHERE session_id = s.id AND member_id = $1) AS user_rsvped
		 FROM space_sessions s
		 LEFT JOIN rsvps r ON r.session_id = s.id
		 %s
		 GROUP BY s.id
		 ORDER BY s.date ASC, s.start_time ASC`, where,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.SpaceSession
	for rows.Next() {
		var s model.SpaceSession
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt, &s.RSVPCount, &s.UserRSVPed); err != nil {
			return nil, fmt.Errorf("scanning session: %w", err)
		}
		sessions = append(sessions, s)
	}

	if sessions == nil {
		sessions = []model.SpaceSession{}
	}

	return sessions, nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID, memberID *uuid.UUID) (*model.SpaceSession, error) {
	var memberIDArg any
	if memberID != nil {
		memberIDArg = *memberID
	} else {
		memberIDArg = uuid.Nil
	}

	var s model.SpaceSession
	err := r.pool.QueryRow(ctx,
		`SELECT s.id, s.title, s.description, s.date::text, to_char(s.start_time, 'HH24:MI'), to_char(s.end_time, 'HH24:MI'), s.status,
		        s.image_url, s.location, s.series_id, s.created_by, s.created_at, s.updated_at,
		        COALESCE(COUNT(r.id), 0) AS rsvp_count,
		        EXISTS(SELECT 1 FROM rsvps WHERE session_id = s.id AND member_id = $2) AS user_rsvped
		 FROM space_sessions s
		 LEFT JOIN rsvps r ON r.session_id = s.id
		 WHERE s.id = $1
		 GROUP BY s.id`, id, memberIDArg,
	).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt, &s.RSVPCount, &s.UserRSVPed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting session by id: %w", err)
	}
	return &s, nil
}

func (r *SessionRepository) Update(ctx context.Context, id uuid.UUID, req model.UpdateSessionRequest, newStatus *string) (*model.SpaceSession, error) {
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Date != nil {
		setClauses = append(setClauses, fmt.Sprintf("date = $%d", argIdx))
		args = append(args, *req.Date)
		argIdx++
	}
	if req.StartTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("start_time = $%d", argIdx))
		args = append(args, *req.StartTime)
		argIdx++
	}
	if req.EndTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("end_time = $%d", argIdx))
		args = append(args, *req.EndTime)
		argIdx++
	}
	if req.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *req.Location)
		argIdx++
	}
	if newStatus != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *newStatus)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id, nil)
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE space_sessions SET %s WHERE id = $%d
		 RETURNING id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), status, image_url, location, series_id, created_by, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var s model.SpaceSession
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("updating session: %w", err)
	}
	return &s, nil
}

func (r *SessionRepository) Cancel(ctx context.Context, id uuid.UUID) (*model.SpaceSession, error) {
	var s model.SpaceSession
	err := r.pool.QueryRow(ctx,
		`UPDATE space_sessions SET status = 'canceled', updated_at = now()
		 WHERE id = $1
		 RETURNING id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), status, image_url, location, series_id, created_by, created_at, updated_at`, id,
	).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("canceling session: %w", err)
	}
	return &s, nil
}

func (r *SessionRepository) UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL *string) (*model.SpaceSession, error) {
	var s model.SpaceSession
	err := r.pool.QueryRow(ctx,
		`UPDATE space_sessions SET image_url = $1, updated_at = now()
		 WHERE id = $2
		 RETURNING id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), status, image_url, location, series_id, created_by, created_at, updated_at`,
		imageURL, id,
	).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("updating session image url: %w", err)
	}
	return &s, nil
}

func (r *SessionRepository) GetRSVPCount(ctx context.Context, sessionID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM rsvps WHERE session_id = $1`, sessionID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("getting rsvp count: %w", err)
	}
	return count, nil
}

func (r *SessionRepository) ListFutureBySeriesID(ctx context.Context, seriesID uuid.UUID) ([]model.SpaceSession, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.title, s.description, s.date::text, to_char(s.start_time, 'HH24:MI'), to_char(s.end_time, 'HH24:MI'), s.status,
		        s.image_url, s.location, s.series_id, s.created_by, s.created_at, s.updated_at,
		        COALESCE(COUNT(r.id), 0) AS rsvp_count
		 FROM space_sessions s
		 LEFT JOIN rsvps r ON r.session_id = s.id
		 WHERE s.series_id = $1 AND s.date >= CURRENT_DATE AND s.status != 'canceled'
		 GROUP BY s.id
		 ORDER BY s.date ASC`, seriesID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing future sessions by series: %w", err)
	}
	defer rows.Close()

	var sessions []model.SpaceSession
	for rows.Next() {
		var s model.SpaceSession
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt, &s.RSVPCount); err != nil {
			return nil, fmt.Errorf("scanning future session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []model.SpaceSession{}
	}
	return sessions, nil
}

func (r *SessionRepository) UpdateBulkBySeriesID(ctx context.Context, seriesID uuid.UUID, req model.UpdateSessionRequest, imageURL *string) (int64, error) {
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.StartTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("start_time = $%d", argIdx))
		args = append(args, *req.StartTime)
		argIdx++
	}
	if req.EndTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("end_time = $%d", argIdx))
		args = append(args, *req.EndTime)
		argIdx++
	}
	if req.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *req.Location)
		argIdx++
	}
	if imageURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", argIdx))
		args = append(args, *imageURL)
		argIdx++
	}

	if len(setClauses) == 0 {
		return 0, nil
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, seriesID)

	query := fmt.Sprintf(
		`UPDATE space_sessions SET %s WHERE series_id = $%d AND date >= CURRENT_DATE AND status != 'canceled'`,
		strings.Join(setClauses, ", "), argIdx,
	)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk updating sessions by series: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *SessionRepository) CancelFutureBySeriesID(ctx context.Context, seriesID uuid.UUID) ([]model.SpaceSession, error) {
	rows, err := r.pool.Query(ctx,
		`UPDATE space_sessions SET status = 'canceled', updated_at = now()
		 WHERE series_id = $1 AND date >= CURRENT_DATE AND status != 'canceled'
		 RETURNING id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), status, image_url, location, series_id, created_by, created_at, updated_at`,
		seriesID,
	)
	if err != nil {
		return nil, fmt.Errorf("canceling future sessions by series: %w", err)
	}
	defer rows.Close()

	var sessions []model.SpaceSession
	for rows.Next() {
		var s model.SpaceSession
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning canceled session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []model.SpaceSession{}
	}
	return sessions, nil
}
