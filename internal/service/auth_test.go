package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

// --- Mock implementations ---

type mockTokenRepo struct {
	tokens       []*model.MagicToken
	recentCount  int
	createCalled bool
	markUsedID   uuid.UUID
	cleanCalled  bool
}

func (m *mockTokenRepo) Create(ctx context.Context, memberID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.MagicToken, error) {
	m.createCalled = true
	t := &model.MagicToken{
		ID:        uuid.New(),
		MemberID:  memberID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	m.tokens = append(m.tokens, t)
	return t, nil
}

func (m *mockTokenRepo) FindValidByHash(ctx context.Context, tokenHash string) (*model.MagicToken, error) {
	for _, t := range m.tokens {
		if t.TokenHash == tokenHash && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockTokenRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	m.markUsedID = id
	for _, t := range m.tokens {
		if t.ID == id {
			now := time.Now()
			t.UsedAt = &now
		}
	}
	return nil
}

func (m *mockTokenRepo) CountRecentByEmail(ctx context.Context, email string) (int, error) {
	return m.recentCount, nil
}

func (m *mockTokenRepo) CleanExpired(ctx context.Context) (int64, error) {
	m.cleanCalled = true
	return 0, nil
}

type mockMagicLinkSender struct {
	lastEmail string
	lastLink  string
	called    bool
}

func (m *mockMagicLinkSender) SendMagicLink(ctx context.Context, toEmail, toName, link string) error {
	m.called = true
	m.lastEmail = toEmail
	m.lastLink = link
	return nil
}

type mockAuthMemberRepo struct {
	members map[uuid.UUID]*model.Member
	byEmail map[string]*model.Member
}

func newMockAuthMemberRepo() *mockAuthMemberRepo {
	return &mockAuthMemberRepo{
		members: make(map[uuid.UUID]*model.Member),
		byEmail: make(map[string]*model.Member),
	}
}

func (m *mockAuthMemberRepo) addMember(member *model.Member) {
	m.members[member.ID] = member
	m.byEmail[member.Email] = member
}

func (m *mockAuthMemberRepo) Create(ctx context.Context, req model.CreateMemberRequest) (*model.Member, error) {
	return nil, nil
}
func (m *mockAuthMemberRepo) List(ctx context.Context, activeFilter string) ([]model.Member, error) {
	return nil, nil
}
func (m *mockAuthMemberRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	return member, nil
}
func (m *mockAuthMemberRepo) GetByEmail(ctx context.Context, email string) (*model.Member, error) {
	member, ok := m.byEmail[email]
	if !ok {
		return nil, nil
	}
	return member, nil
}
func (m *mockAuthMemberRepo) Update(ctx context.Context, id uuid.UUID, req model.UpdateMemberRequest) (*model.Member, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	if req.Name != nil {
		member.Name = *req.Name
	}
	if req.TelegramHandle != nil {
		member.TelegramHandle = req.TelegramHandle
	}
	return member, nil
}
func (m *mockAuthMemberRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockAuthMemberRepo) HasRSVPs(ctx context.Context, memberID uuid.UUID) (bool, error) {
	return false, nil
}

func newTestAuthService() (*AuthService, *mockTokenRepo, *mockAuthMemberRepo, *mockMagicLinkSender) {
	tokenRepo := &mockTokenRepo{}
	memberRepo := newMockAuthMemberRepo()
	emailSender := &mockMagicLinkSender{}
	svc := NewAuthService(tokenRepo, memberRepo, emailSender, "test-secret-key", "http://localhost:5173", false)
	return svc, tokenRepo, memberRepo, emailSender
}

// --- Tests ---

func TestGenerateToken(t *testing.T) {
	raw, hash, err := generateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if raw == "" {
		t.Error("raw token should not be empty")
	}
	if hash == "" {
		t.Error("hash should not be empty")
	}

	// Verify hash matches
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
	if hash != expected {
		t.Errorf("hash mismatch: got %s, want %s", hash, expected)
	}
}

func TestHashToken(t *testing.T) {
	hash1 := hashToken("test-token")
	hash2 := hashToken("test-token")
	hash3 := hashToken("different-token")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestRequestMagicLink_ActiveMember(t *testing.T) {
	svc, tokenRepo, memberRepo, _ := newTestAuthService()

	member := &model.Member{
		ID:       uuid.New(),
		Email:    "jane@example.com",
		Name:     "Jane",
		IsActive: true,
	}
	memberRepo.addMember(member)

	err := svc.RequestMagicLink(context.Background(), "jane@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !tokenRepo.createCalled {
		t.Error("expected token to be created")
	}
}

func TestRequestMagicLink_UnknownEmail_NoError(t *testing.T) {
	svc, tokenRepo, _, _ := newTestAuthService()

	err := svc.RequestMagicLink(context.Background(), "unknown@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenRepo.createCalled {
		t.Error("should not create token for unknown email")
	}
}

func TestRequestMagicLink_InactiveMember_NoError(t *testing.T) {
	svc, tokenRepo, memberRepo, _ := newTestAuthService()

	member := &model.Member{
		ID:       uuid.New(),
		Email:    "inactive@example.com",
		Name:     "Inactive",
		IsActive: false,
	}
	memberRepo.addMember(member)

	err := svc.RequestMagicLink(context.Background(), "inactive@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenRepo.createCalled {
		t.Error("should not create token for inactive member")
	}
}

func TestRequestMagicLink_RateLimited(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()
	svc.tokenRepo.(*mockTokenRepo).recentCount = 5

	member := &model.Member{
		ID:       uuid.New(),
		Email:    "jane@example.com",
		Name:     "Jane",
		IsActive: true,
	}
	memberRepo.addMember(member)

	err := svc.RequestMagicLink(context.Background(), "jane@example.com")
	if err != ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestRequestMagicLink_EmptyEmail(t *testing.T) {
	svc, tokenRepo, _, _ := newTestAuthService()

	err := svc.RequestMagicLink(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenRepo.createCalled {
		t.Error("should not create token for empty email")
	}
}

func TestVerifyToken_Valid(t *testing.T) {
	svc, tokenRepo, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	member := &model.Member{
		ID:       memberID,
		Email:    "jane@example.com",
		Name:     "Jane",
		IsActive: true,
	}
	memberRepo.addMember(member)

	rawToken := "test-raw-token-123"
	hash := hashToken(rawToken)
	tokenRepo.tokens = append(tokenRepo.tokens, &model.MagicToken{
		ID:        uuid.New(),
		MemberID:  memberID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})

	result, err := svc.VerifyToken(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != memberID {
		t.Errorf("expected member ID %s, got %s", memberID, result.ID)
	}
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	svc, _, _, _ := newTestAuthService()

	_, err := svc.VerifyToken(context.Background(), "nonexistent-token")
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	svc, tokenRepo, _, _ := newTestAuthService()

	rawToken := "expired-token"
	hash := hashToken(rawToken)
	tokenRepo.tokens = append(tokenRepo.tokens, &model.MagicToken{
		ID:        uuid.New(),
		MemberID:  uuid.New(),
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-1 * time.Minute), // expired
	})

	_, err := svc.VerifyToken(context.Background(), rawToken)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyToken_UsedToken(t *testing.T) {
	svc, tokenRepo, _, _ := newTestAuthService()

	now := time.Now()
	rawToken := "used-token"
	hash := hashToken(rawToken)
	tokenRepo.tokens = append(tokenRepo.tokens, &model.MagicToken{
		ID:        uuid.New(),
		MemberID:  uuid.New(),
		TokenHash: hash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    &now,
	})

	_, err := svc.VerifyToken(context.Background(), rawToken)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyToken_InactiveMember(t *testing.T) {
	svc, tokenRepo, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	member := &model.Member{
		ID:       memberID,
		Email:    "inactive@example.com",
		Name:     "Inactive",
		IsActive: false,
	}
	memberRepo.addMember(member)

	rawToken := "inactive-member-token"
	hash := hashToken(rawToken)
	tokenRepo.tokens = append(tokenRepo.tokens, &model.MagicToken{
		ID:        uuid.New(),
		MemberID:  memberID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})

	_, err := svc.VerifyToken(context.Background(), rawToken)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestCookieEncodeAndDecode(t *testing.T) {
	svc, _, _, _ := newTestAuthService()

	memberID := uuid.New()
	cookie, err := svc.CreateSessionCookie(memberID)
	if err != nil {
		t.Fatalf("unexpected error creating cookie: %v", err)
	}

	if cookie.Name != "session" {
		t.Errorf("expected cookie name 'session', got %s", cookie.Name)
	}
	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
	if cookie.Secure {
		t.Error("cookie should not be Secure for http://localhost")
	}
	if cookie.SameSite != 2 { // http.SameSiteLaxMode
		t.Errorf("expected SameSiteLax, got %d", cookie.SameSite)
	}

	decoded, err := svc.ValidateSessionCookie(cookie.Value)
	if err != nil {
		t.Fatalf("unexpected error decoding cookie: %v", err)
	}
	if decoded != memberID {
		t.Errorf("expected member ID %s, got %s", memberID, decoded)
	}
}

func TestCookieInvalidValue(t *testing.T) {
	svc, _, _, _ := newTestAuthService()

	_, err := svc.ValidateSessionCookie("invalid-cookie-value")
	if err == nil {
		t.Error("expected error for invalid cookie value")
	}
}

func TestClearSessionCookie(t *testing.T) {
	svc, _, _, _ := newTestAuthService()

	cookie := svc.ClearSessionCookie()
	if cookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", cookie.MaxAge)
	}
	if cookie.Value != "" {
		t.Error("expected empty cookie value")
	}
}

func TestUpdateProfile_ValidName(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	member := &model.Member{
		ID:       memberID,
		Email:    "jane@example.com",
		Name:     "Jane Doe",
		IsActive: true,
	}
	memberRepo.addMember(member)

	name := "Jane Smith"
	updated, err := svc.UpdateProfile(context.Background(), memberID, &name, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Jane Smith" {
		t.Errorf("expected name 'Jane Smith', got %s", updated.Name)
	}
}

func TestUpdateProfile_EmptyName(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	name := ""
	_, err := svc.UpdateProfile(context.Background(), memberID, &name, nil)
	var ve *ValidationError
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !isValidationError(err, &ve) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if ve.Details["name"] != "required" {
		t.Errorf("expected name required, got %s", ve.Details["name"])
	}
}

func TestUpdateProfile_StripTelegramAt(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	handle := "@janedoe"
	updated, err := svc.UpdateProfile(context.Background(), memberID, nil, &handle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.TelegramHandle == nil || *updated.TelegramHandle != "janedoe" {
		t.Error("expected telegram handle 'janedoe' with @ stripped")
	}
}

func TestCleanExpiredTokens(t *testing.T) {
	svc, tokenRepo, _, _ := newTestAuthService()

	svc.CleanExpiredTokens(context.Background())

	if !tokenRepo.cleanCalled {
		t.Error("expected clean to be called")
	}
}

func isValidationError(err error, target **ValidationError) bool {
	if ve, ok := err.(*ValidationError); ok {
		*target = ve
		return true
	}
	return false
}
