package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"

	"github.com/developer-space/api/internal/model"
)

var (
	ErrInvalidToken  = errors.New("invalid or expired login link")
	ErrInactiveMember = errors.New("member is inactive")
	ErrRateLimited   = errors.New("too many requests")
)

const (
	tokenExpiry    = 15 * time.Minute
	cookieName     = "session"
	cookieMaxAge   = 30 * 24 * 60 * 60 // 30 days in seconds
	rateLimitMax   = 5
)

// MagicTokenRepo defines the data access interface for magic tokens.
type MagicTokenRepo interface {
	Create(ctx context.Context, memberID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.MagicToken, error)
	FindValidByHash(ctx context.Context, tokenHash string) (*model.MagicToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	CountRecentByEmail(ctx context.Context, email string) (int, error)
	CleanExpired(ctx context.Context) (int64, error)
}

// MagicLinkSender sends magic link emails.
type MagicLinkSender interface {
	SendMagicLink(ctx context.Context, toEmail, toName, link string) error
}

type AuthService struct {
	tokenRepo    MagicTokenRepo
	memberRepo   MemberRepo
	emailSender  MagicLinkSender
	sc           *securecookie.SecureCookie
	frontendURL  string
	isSecure     bool
}

func NewAuthService(tokenRepo MagicTokenRepo, memberRepo MemberRepo, emailSender MagicLinkSender, sessionSecret string, frontendURL string, isSecure bool) *AuthService {
	hashKey := sha256.Sum256([]byte(sessionSecret))
	sc := securecookie.New(hashKey[:], nil)
	sc.MaxAge(cookieMaxAge)

	return &AuthService{
		tokenRepo:   tokenRepo,
		memberRepo:  memberRepo,
		emailSender: emailSender,
		sc:          sc,
		frontendURL: frontendURL,
		isSecure:    isSecure,
	}
}

// RequestMagicLink generates a token and sends a magic link email.
// Always returns nil to prevent user enumeration — errors are logged, not returned.
func (s *AuthService) RequestMagicLink(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}

	// Check rate limit
	count, err := s.tokenRepo.CountRecentByEmail(ctx, email)
	if err != nil {
		slog.Error("failed to check rate limit", "error", err)
		return nil
	}
	if count >= rateLimitMax {
		return ErrRateLimited
	}

	// Look up active member
	member, err := s.memberRepo.GetByEmail(ctx, email)
	if err != nil {
		slog.Error("failed to look up member", "error", err)
		return nil
	}
	if member == nil || !member.IsActive {
		return nil // no enumeration
	}

	// Generate token
	rawToken, tokenHash, err := generateToken()
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		return nil
	}

	expiresAt := time.Now().Add(tokenExpiry)
	if _, err := s.tokenRepo.Create(ctx, member.ID, tokenHash, expiresAt); err != nil {
		slog.Error("failed to store token", "error", err)
		return nil
	}

	// Send email (fire-and-forget)
	link := fmt.Sprintf("%s/auth/verify?token=%s", s.frontendURL, rawToken)
	go func() {
		if err := s.emailSender.SendMagicLink(context.Background(), member.Email, member.Name, link); err != nil {
			slog.Warn("failed to send magic link email", "email", email, "error", err)
		}
	}()

	return nil
}

// VerifyToken validates a magic link token and returns the member.
func (s *AuthService) VerifyToken(ctx context.Context, rawToken string) (*model.Member, error) {
	tokenHash := hashToken(rawToken)

	token, err := s.tokenRepo.FindValidByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("looking up token: %w", err)
	}
	if token == nil {
		return nil, ErrInvalidToken
	}

	// Mark used immediately
	if err := s.tokenRepo.MarkUsed(ctx, token.ID); err != nil {
		return nil, fmt.Errorf("marking token used: %w", err)
	}

	// Look up member
	member, err := s.memberRepo.GetByID(ctx, token.MemberID)
	if err != nil {
		return nil, fmt.Errorf("looking up member: %w", err)
	}
	if member == nil || !member.IsActive {
		return nil, ErrInvalidToken
	}

	return member, nil
}

// CreateSessionCookie creates an encoded session cookie value.
func (s *AuthService) CreateSessionCookie(memberID uuid.UUID) (*http.Cookie, error) {
	value := map[string]string{
		"member_id": memberID.String(),
	}

	encoded, err := s.sc.Encode(cookieName, value)
	if err != nil {
		return nil, fmt.Errorf("encoding cookie: %w", err)
	}

	return &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   s.isSecure,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// ValidateSessionCookie decodes a session cookie and returns the member ID.
func (s *AuthService) ValidateSessionCookie(cookieValue string) (uuid.UUID, error) {
	value := map[string]string{}
	if err := s.sc.Decode(cookieName, cookieValue, &value); err != nil {
		return uuid.Nil, fmt.Errorf("decoding cookie: %w", err)
	}

	memberID, err := uuid.Parse(value["member_id"])
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing member_id: %w", err)
	}

	return memberID, nil
}

// ClearSessionCookie returns a cookie that clears the session.
func (s *AuthService) ClearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.isSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ProfileUpdateInput contains the fields a member can update on their own profile.
type ProfileUpdateInput struct {
	Name            *string  `json:"name"`
	TelegramHandle  *string  `json:"telegram_handle"`
	Bio             *string  `json:"bio"`
	Skills          []string `json:"skills"`
	LinkedinURL     *string  `json:"linkedin_url"`
	InstagramHandle *string  `json:"instagram_handle"`
	GithubUsername  *string  `json:"github_username"`
}

// UpdateProfile updates the current user's profile fields.
func (s *AuthService) UpdateProfile(ctx context.Context, memberID uuid.UUID, input ProfileUpdateInput) (*model.Member, error) {
	// Validate name if provided
	if input.Name != nil {
		n := strings.TrimSpace(*input.Name)
		if n == "" {
			return nil, &ValidationError{Details: map[string]string{"name": "required"}}
		}
		input.Name = &n
	}

	// Strip @ from telegram handle
	if input.TelegramHandle != nil {
		h := strings.TrimPrefix(strings.TrimSpace(*input.TelegramHandle), "@")
		input.TelegramHandle = &h
	}

	// Validate bio length
	if input.Bio != nil {
		b := strings.TrimSpace(*input.Bio)
		if len(b) > 500 {
			return nil, &ValidationError{Details: map[string]string{"bio": "must be 500 characters or fewer"}}
		}
		input.Bio = &b
	}

	// Validate and normalize skills
	if input.Skills != nil {
		if len(input.Skills) > 10 {
			return nil, &ValidationError{Details: map[string]string{"skills": "maximum 10 tags allowed"}}
		}
		normalized := make([]string, 0, len(input.Skills))
		for _, skill := range input.Skills {
			s := strings.ToLower(strings.TrimSpace(skill))
			if s != "" {
				normalized = append(normalized, s)
			}
		}
		input.Skills = normalized
	}

	// Validate linkedin_url
	if input.LinkedinURL != nil {
		u := strings.TrimSpace(*input.LinkedinURL)
		if u != "" {
			if _, err := url.ParseRequestURI(u); err != nil || (!strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://")) {
				return nil, &ValidationError{Details: map[string]string{"linkedin_url": "must be a valid URL"}}
			}
		}
		input.LinkedinURL = &u
	}

	// Strip @ from instagram handle
	if input.InstagramHandle != nil {
		h := strings.TrimPrefix(strings.TrimSpace(*input.InstagramHandle), "@")
		input.InstagramHandle = &h
	}

	// Strip @ from github username
	if input.GithubUsername != nil {
		h := strings.TrimPrefix(strings.TrimSpace(*input.GithubUsername), "@")
		input.GithubUsername = &h
	}

	req := model.UpdateMemberRequest{
		Name:            input.Name,
		TelegramHandle:  input.TelegramHandle,
		Bio:             input.Bio,
		Skills:          input.Skills,
		LinkedinURL:     input.LinkedinURL,
		InstagramHandle: input.InstagramHandle,
		GithubUsername:  input.GithubUsername,
	}

	member, err := s.memberRepo.Update(ctx, memberID, req)
	if err != nil {
		return nil, fmt.Errorf("updating profile: %w", err)
	}
	if member == nil {
		return nil, ErrNotFound
	}

	return member, nil
}

// GetPublicProfile returns publicly visible profile information for a member.
func (s *AuthService) GetPublicProfile(ctx context.Context, memberID uuid.UUID) (*model.PublicMember, error) {
	member, err := s.memberRepo.GetByIDPublic(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("getting public profile: %w", err)
	}
	if member == nil {
		return nil, ErrNotFound
	}
	return member, nil
}

// CleanExpiredTokens removes expired and used tokens.
func (s *AuthService) CleanExpiredTokens(ctx context.Context) {
	deleted, err := s.tokenRepo.CleanExpired(ctx)
	if err != nil {
		slog.Warn("failed to clean expired tokens", "error", err)
		return
	}
	if deleted > 0 {
		slog.Debug("cleaned expired tokens", "count", deleted)
	}
}

// generateToken creates a cryptographically random token and its SHA-256 hash.
func generateToken() (rawToken string, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating random bytes: %w", err)
	}
	raw := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	return raw, hashToken(raw), nil
}

// hashToken computes the SHA-256 hash of a token string.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
