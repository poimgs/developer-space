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
func (m *mockAuthMemberRepo) GetByIDPublic(ctx context.Context, id uuid.UUID) (*model.PublicMember, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	if !member.IsActive {
		return nil, nil
	}
	return &model.PublicMember{
		ID:              member.ID,
		Name:            member.Name,
		TelegramHandle:  member.TelegramHandle,
		Bio:             member.Bio,
		Skills:          member.Skills,
		LinkedinURL:     member.LinkedinURL,
		InstagramHandle: member.InstagramHandle,
		GithubUsername:  member.GithubUsername,
	}, nil
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
	if req.Bio != nil {
		member.Bio = req.Bio
	}
	if req.Skills != nil {
		member.Skills = req.Skills
	}
	if req.LinkedinURL != nil {
		member.LinkedinURL = req.LinkedinURL
	}
	if req.InstagramHandle != nil {
		member.InstagramHandle = req.InstagramHandle
	}
	if req.GithubUsername != nil {
		member.GithubUsername = req.GithubUsername
	}
	return member, nil
}
func (m *mockAuthMemberRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockAuthMemberRepo) HasRSVPs(ctx context.Context, memberID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockAuthMemberRepo) DistinctSkills(ctx context.Context) ([]string, error) {
	return []string{}, nil
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
	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{Name: &name})
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
	_, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{Name: &name})
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
	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{TelegramHandle: &handle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.TelegramHandle == nil || *updated.TelegramHandle != "janedoe" {
		t.Error("expected telegram handle 'janedoe' with @ stripped")
	}
}

func TestUpdateProfile_Bio(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	bio := "I'm a software developer who loves Go."
	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{Bio: &bio})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Bio == nil || *updated.Bio != bio {
		t.Errorf("expected bio %q, got %v", bio, updated.Bio)
	}
}

func TestUpdateProfile_BioTooLong(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	// Create a 501-char string
	bio := ""
	for i := 0; i < 501; i++ {
		bio += "a"
	}
	_, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{Bio: &bio})
	var ve *ValidationError
	if err == nil {
		t.Fatal("expected error for bio > 500 chars")
	}
	if !isValidationError(err, &ve) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if ve.Details["bio"] != "must be 500 characters or fewer" {
		t.Errorf("expected bio validation message, got %s", ve.Details["bio"])
	}
}

func TestUpdateProfile_Skills(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	skills := []string{" Go ", "REACT", "TypeScript"}
	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{Skills: skills})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be trimmed and lowercased
	if len(updated.Skills) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(updated.Skills))
	}
	if updated.Skills[0] != "go" {
		t.Errorf("expected 'go', got %q", updated.Skills[0])
	}
	if updated.Skills[1] != "react" {
		t.Errorf("expected 'react', got %q", updated.Skills[1])
	}
}

func TestUpdateProfile_SkillsTooMany(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	skills := make([]string, 11)
	for i := range skills {
		skills[i] = fmt.Sprintf("skill%d", i)
	}
	_, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{Skills: skills})
	var ve *ValidationError
	if err == nil {
		t.Fatal("expected error for skills > 10")
	}
	if !isValidationError(err, &ve) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if ve.Details["skills"] != "maximum 10 tags allowed" {
		t.Errorf("expected skills validation message, got %s", ve.Details["skills"])
	}
}

func TestUpdateProfile_LinkedinURLValid(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	url := "https://linkedin.com/in/janedoe"
	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{LinkedinURL: &url})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.LinkedinURL == nil || *updated.LinkedinURL != url {
		t.Errorf("expected linkedin_url %q, got %v", url, updated.LinkedinURL)
	}
}

func TestUpdateProfile_LinkedinURLInvalid(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	url := "not-a-url"
	_, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{LinkedinURL: &url})
	var ve *ValidationError
	if err == nil {
		t.Fatal("expected error for invalid linkedin_url")
	}
	if !isValidationError(err, &ve) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if ve.Details["linkedin_url"] != "must be a valid URL" {
		t.Errorf("expected linkedin_url validation message, got %s", ve.Details["linkedin_url"])
	}
}

func TestUpdateProfile_StripInstagramAt(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	handle := "@janedoe"
	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{InstagramHandle: &handle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.InstagramHandle == nil || *updated.InstagramHandle != "janedoe" {
		t.Errorf("expected instagram handle 'janedoe', got %v", updated.InstagramHandle)
	}
}

func TestUpdateProfile_StripGithubAt(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true})

	username := "@janedoe"
	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{GithubUsername: &username})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.GithubUsername == nil || *updated.GithubUsername != "janedoe" {
		t.Errorf("expected github username 'janedoe', got %v", updated.GithubUsername)
	}
}

func TestUpdateProfile_AllFieldsTogether(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane Doe", IsActive: true})

	name := "Jane Smith"
	telegram := "@janesmith"
	bio := "Developer and writer"
	skills := []string{"Go", "React"}
	linkedin := "https://linkedin.com/in/janesmith"
	instagram := "@jane_ig"
	github := "@janegh"

	updated, err := svc.UpdateProfile(context.Background(), memberID, ProfileUpdateInput{
		Name:            &name,
		TelegramHandle:  &telegram,
		Bio:             &bio,
		Skills:          skills,
		LinkedinURL:     &linkedin,
		InstagramHandle: &instagram,
		GithubUsername:  &github,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Jane Smith" {
		t.Errorf("expected name 'Jane Smith', got %s", updated.Name)
	}
	if updated.TelegramHandle == nil || *updated.TelegramHandle != "janesmith" {
		t.Errorf("expected telegram 'janesmith', got %v", updated.TelegramHandle)
	}
	if updated.Bio == nil || *updated.Bio != "Developer and writer" {
		t.Errorf("expected bio, got %v", updated.Bio)
	}
	if len(updated.Skills) != 2 || updated.Skills[0] != "go" {
		t.Errorf("expected lowercase skills, got %v", updated.Skills)
	}
	if updated.LinkedinURL == nil || *updated.LinkedinURL != linkedin {
		t.Errorf("expected linkedin URL, got %v", updated.LinkedinURL)
	}
	if updated.InstagramHandle == nil || *updated.InstagramHandle != "jane_ig" {
		t.Errorf("expected instagram 'jane_ig', got %v", updated.InstagramHandle)
	}
	if updated.GithubUsername == nil || *updated.GithubUsername != "janegh" {
		t.Errorf("expected github 'janegh', got %v", updated.GithubUsername)
	}
}

func TestGetPublicProfile_Found(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	bio := "Hello world"
	memberRepo.addMember(&model.Member{
		ID:       memberID,
		Email:    "jane@example.com",
		Name:     "Jane Doe",
		IsActive: true,
		Bio:      &bio,
		Skills:   []string{"go", "react"},
	})

	profile, err := svc.GetPublicProfile(context.Background(), memberID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ID != memberID {
		t.Errorf("expected ID %s, got %s", memberID, profile.ID)
	}
	if profile.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got %s", profile.Name)
	}
	if profile.Bio == nil || *profile.Bio != "Hello world" {
		t.Errorf("expected bio 'Hello world', got %v", profile.Bio)
	}
}

func TestGetPublicProfile_NotFound(t *testing.T) {
	svc, _, _, _ := newTestAuthService()

	_, err := svc.GetPublicProfile(context.Background(), uuid.New())
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetPublicProfile_InactiveMember(t *testing.T) {
	svc, _, memberRepo, _ := newTestAuthService()

	memberID := uuid.New()
	memberRepo.addMember(&model.Member{
		ID:       memberID,
		Email:    "inactive@example.com",
		Name:     "Inactive",
		IsActive: false,
	})

	_, err := svc.GetPublicProfile(context.Background(), memberID)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for inactive member, got %v", err)
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
