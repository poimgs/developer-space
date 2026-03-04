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

type SeriesRepository struct {
	pool *pgxpool.Pool
}

func NewSeriesRepository(pool *pgxpool.Pool) *SeriesRepository {
	return &SeriesRepository{pool: pool}
}

func (r *SeriesRepository) Create(ctx context.Context, series model.SessionSeries) (*model.SessionSeries, error) {
	var s model.SessionSeries
	err := r.pool.QueryRow(ctx,
		`INSERT INTO session_series (title, description, day_of_week, start_time, end_time, image_url, location, every_n_weeks, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, title, description, day_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), image_url, location, every_n_weeks, is_active, created_by, created_at, updated_at`,
		series.Title, series.Description, series.DayOfWeek, series.StartTime, series.EndTime, series.ImageURL, series.Location, series.EveryNWeeks, series.CreatedBy,
	).Scan(&s.ID, &s.Title, &s.Description, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ImageURL, &s.Location, &s.EveryNWeeks, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting session series: %w", err)
	}
	return &s, nil
}

func (r *SeriesRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SessionSeries, error) {
	var s model.SessionSeries
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, description, day_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), image_url, location, every_n_weeks, is_active, created_by, created_at, updated_at
		 FROM session_series WHERE id = $1`, id,
	).Scan(&s.ID, &s.Title, &s.Description, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ImageURL, &s.Location, &s.EveryNWeeks, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting session series: %w", err)
	}
	return &s, nil
}

func (r *SeriesRepository) ListActive(ctx context.Context) ([]model.SessionSeries, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, description, day_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), image_url, location, every_n_weeks, is_active, created_by, created_at, updated_at
		 FROM session_series WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("listing active series: %w", err)
	}
	defer rows.Close()

	var series []model.SessionSeries
	for rows.Next() {
		var s model.SessionSeries
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ImageURL, &s.Location, &s.EveryNWeeks, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning session series: %w", err)
		}
		series = append(series, s)
	}

	if series == nil {
		series = []model.SessionSeries{}
	}
	return series, nil
}

func (r *SeriesRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE session_series SET is_active = false, updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deactivating session series: %w", err)
	}
	return nil
}

// GenerateSessions creates sessions for the given series for each matching weekday
// between fromDate and toDate (inclusive). Uses ON CONFLICT DO NOTHING for idempotency.
func (r *SeriesRepository) GenerateSessions(ctx context.Context, series model.SessionSeries, fromDate, toDate time.Time) ([]model.SpaceSession, error) {
	// Determine the interval in weeks (default to 1)
	interval := series.EveryNWeeks
	if interval < 1 {
		interval = 1
	}
	advanceDays := 7 * interval

	// Find the first matching weekday >= fromDate
	current := fromDate
	targetDay := time.Weekday(series.DayOfWeek)
	for current.Weekday() != targetDay {
		current = current.AddDate(0, 0, 1)
	}

	// For bi-weekly+ series, ensure phase alignment by finding the last existing session
	if interval > 1 {
		var lastDate *time.Time
		err := r.pool.QueryRow(ctx,
			`SELECT MAX(date) FROM space_sessions WHERE series_id = $1`, series.ID,
		).Scan(&lastDate)
		if err == nil && lastDate != nil {
			// Advance from the last existing session by the interval
			candidate := lastDate.AddDate(0, 0, advanceDays)
			if candidate.After(current) {
				current = candidate
			}
		}
	}

	var sessions []model.SpaceSession
	for !current.After(toDate) {
		dateStr := current.Format("2006-01-02")
		var s model.SpaceSession
		err := r.pool.QueryRow(ctx,
			`INSERT INTO space_sessions (title, description, date, start_time, end_time, image_url, location, status, series_id, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, 'scheduled', $8, $9)
			 ON CONFLICT (series_id, date) WHERE series_id IS NOT NULL DO NOTHING
			 RETURNING id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), status, image_url, location, series_id, created_by, created_at, updated_at`,
			series.Title, series.Description, dateStr, series.StartTime, series.EndTime, series.ImageURL, series.Location, series.ID, series.CreatedBy,
		).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Status, &s.ImageURL, &s.Location, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// ON CONFLICT DO NOTHING — session already exists for this date
				current = current.AddDate(0, 0, advanceDays)
				continue
			}
			return nil, fmt.Errorf("generating session for %s: %w", dateStr, err)
		}
		sessions = append(sessions, s)
		current = current.AddDate(0, 0, advanceDays)
	}

	return sessions, nil
}

func (r *SeriesRepository) Update(ctx context.Context, id uuid.UUID, req model.UpdateSeriesRequest) (*model.SessionSeries, error) {
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
	if req.ImageURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", argIdx))
		args = append(args, *req.ImageURL)
		argIdx++
	}
	if req.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *req.Location)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE session_series SET %s WHERE id = $%d
		 RETURNING id, title, description, day_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), image_url, location, every_n_weeks, is_active, created_by, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var s model.SessionSeries
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&s.ID, &s.Title, &s.Description, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ImageURL, &s.Location, &s.EveryNWeeks, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("updating session series: %w", err)
	}
	return &s, nil
}

func (r *SeriesRepository) UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL *string) (*model.SessionSeries, error) {
	var s model.SessionSeries
	err := r.pool.QueryRow(ctx,
		`UPDATE session_series SET image_url = $1, updated_at = now()
		 WHERE id = $2
		 RETURNING id, title, description, day_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), image_url, location, every_n_weeks, is_active, created_by, created_at, updated_at`,
		imageURL, id,
	).Scan(&s.ID, &s.Title, &s.Description, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ImageURL, &s.Location, &s.EveryNWeeks, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("updating series image url: %w", err)
	}
	return &s, nil
}
