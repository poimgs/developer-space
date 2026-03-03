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

// RSVPTxResult contains the result of the atomic RSVP creation transaction.
type RSVPTxResult struct {
	RSVP    *model.RSVP
	Session *model.SpaceSession
}

// Sentinel errors for RSVP operations.
var (
	ErrRSVPSessionCanceled = errors.New("cannot RSVP to a canceled session")
	ErrRSVPSessionPast     = errors.New("cannot RSVP to a past session")
	ErrRSVPSessionFull     = errors.New("this session is full")
	ErrRSVPDuplicate       = errors.New("you have already RSVPed to this session")
	ErrRSVPNotFound        = errors.New("you have not RSVPed to this session")
)

type RSVPRepository struct {
	pool *pgxpool.Pool
}

func NewRSVPRepository(pool *pgxpool.Pool) *RSVPRepository {
	return &RSVPRepository{pool: pool}
}

// CreateAtomic performs RSVP creation inside a single transaction with SELECT FOR UPDATE
// to ensure atomic capacity checking. Returns the RSVP and the locked session data.
// Returns nil, nil if the session does not exist.
func (r *RSVPRepository) CreateAtomic(ctx context.Context, sessionID, memberID uuid.UUID) (*RSVPTxResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the session row and fetch it
	var s model.SpaceSession
	err = tx.QueryRow(ctx,
		`SELECT id, title, description, date::text, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), capacity, status, series_id, created_by, created_at, updated_at
		 FROM space_sessions
		 WHERE id = $1
		 FOR UPDATE`, sessionID,
	).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Capacity, &s.Status, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("locking session: %w", err)
	}

	// Check session is not canceled
	if s.Status == "canceled" {
		return nil, ErrRSVPSessionCanceled
	}

	// Check session is not in the past
	today := time.Now().Format("2006-01-02")
	if s.Date < today {
		return nil, ErrRSVPSessionPast
	}

	// Count existing RSVPs
	var count int
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM rsvps WHERE session_id = $1`, sessionID,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("counting rsvps: %w", err)
	}

	// Check capacity
	if count >= s.Capacity {
		return nil, ErrRSVPSessionFull
	}

	// Check for duplicate
	var existingID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT id FROM rsvps WHERE session_id = $1 AND member_id = $2`, sessionID, memberID,
	).Scan(&existingID)
	if err == nil {
		return nil, ErrRSVPDuplicate
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("checking duplicate rsvp: %w", err)
	}

	// Insert RSVP
	var rsvp model.RSVP
	err = tx.QueryRow(ctx,
		`INSERT INTO rsvps (session_id, member_id)
		 VALUES ($1, $2)
		 RETURNING id, session_id, member_id, created_at`,
		sessionID, memberID,
	).Scan(&rsvp.ID, &rsvp.SessionID, &rsvp.MemberID, &rsvp.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting rsvp: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing rsvp: %w", err)
	}

	s.RSVPCount = count + 1
	return &RSVPTxResult{RSVP: &rsvp, Session: &s}, nil
}

// Delete removes an RSVP. Returns the session for notification purposes.
// Returns nil, nil if the session does not exist.
func (r *RSVPRepository) Delete(ctx context.Context, sessionID, memberID uuid.UUID) (*model.SpaceSession, error) {
	session, err := r.getSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Check session is not in the past
	today := time.Now().Format("2006-01-02")
	if session.Date < today {
		return nil, ErrRSVPSessionPast
	}

	tag, err := r.pool.Exec(ctx,
		`DELETE FROM rsvps WHERE session_id = $1 AND member_id = $2`, sessionID, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("deleting rsvp: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrRSVPNotFound
	}

	return session, nil
}

// ListBySession returns all RSVPs for a session with member details, ordered by created_at ASC.
// Returns nil, nil if the session does not exist.
func (r *RSVPRepository) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPWithMember, error) {
	exists, err := r.sessionExists(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx,
		`SELECT r.id, r.session_id, m.id, m.name, m.telegram_handle, r.created_at
		 FROM rsvps r
		 JOIN members m ON m.id = r.member_id
		 WHERE r.session_id = $1
		 ORDER BY r.created_at ASC`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing rsvps: %w", err)
	}
	defer rows.Close()

	var rsvps []model.RSVPWithMember
	for rows.Next() {
		var rv model.RSVPWithMember
		if err := rows.Scan(&rv.ID, &rv.SessionID, &rv.Member.ID, &rv.Member.Name, &rv.Member.TelegramHandle, &rv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning rsvp: %w", err)
		}
		rsvps = append(rsvps, rv)
	}

	if rsvps == nil {
		rsvps = []model.RSVPWithMember{}
	}

	return rsvps, nil
}

// ListEmailsBySession returns the name and email of all RSVPed members for a session.
func (r *RSVPRepository) ListEmailsBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPRecipient, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.name, m.email
		 FROM rsvps r
		 JOIN members m ON m.id = r.member_id
		 WHERE r.session_id = $1`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing rsvp emails: %w", err)
	}
	defer rows.Close()

	var recipients []model.RSVPRecipient
	for rows.Next() {
		var rc model.RSVPRecipient
		if err := rows.Scan(&rc.Name, &rc.Email); err != nil {
			return nil, fmt.Errorf("scanning rsvp email: %w", err)
		}
		recipients = append(recipients, rc)
	}

	return recipients, nil
}

func (r *RSVPRepository) getSession(ctx context.Context, id uuid.UUID) (*model.SpaceSession, error) {
	var s model.SpaceSession
	err := r.pool.QueryRow(ctx,
		`SELECT s.id, s.title, s.description, s.date::text, to_char(s.start_time, 'HH24:MI'), to_char(s.end_time, 'HH24:MI'), s.capacity, s.status,
		        s.series_id, s.created_by, s.created_at, s.updated_at,
		        COALESCE(COUNT(r.id), 0) AS rsvp_count
		 FROM space_sessions s
		 LEFT JOIN rsvps r ON r.session_id = s.id
		 WHERE s.id = $1
		 GROUP BY s.id`, id,
	).Scan(&s.ID, &s.Title, &s.Description, &s.Date, &s.StartTime, &s.EndTime, &s.Capacity, &s.Status, &s.SeriesID, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt, &s.RSVPCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting session: %w", err)
	}
	return &s, nil
}

func (r *RSVPRepository) sessionExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM space_sessions WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking session exists: %w", err)
	}
	return exists, nil
}
