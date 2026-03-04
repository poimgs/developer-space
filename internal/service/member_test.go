package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

// mockMemberRepo implements MemberRepo for testing.
type mockMemberRepo struct {
	members    map[uuid.UUID]*model.Member
	byEmail    map[string]*model.Member
	rsvpCounts map[uuid.UUID]int
	createErr  error
	updateErr  error
	deleteErr  error
}

func newMockRepo() *mockMemberRepo {
	return &mockMemberRepo{
		members:    make(map[uuid.UUID]*model.Member),
		byEmail:    make(map[string]*model.Member),
		rsvpCounts: make(map[uuid.UUID]int),
	}
}

func (m *mockMemberRepo) Create(ctx context.Context, req model.CreateMemberRequest) (*model.Member, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	member := &model.Member{
		ID:             uuid.New(),
		Email:          req.Email,
		Name:           req.Name,
		TelegramHandle: req.TelegramHandle,
		IsAdmin:        req.IsAdmin,
		IsActive:       true,
	}
	m.members[member.ID] = member
	m.byEmail[member.Email] = member
	return member, nil
}

func (m *mockMemberRepo) List(ctx context.Context, activeFilter string) ([]model.Member, error) {
	var result []model.Member
	for _, member := range m.members {
		switch activeFilter {
		case "false":
			if !member.IsActive {
				result = append(result, *member)
			}
		case "all":
			result = append(result, *member)
		default:
			if member.IsActive {
				result = append(result, *member)
			}
		}
	}
	if result == nil {
		result = []model.Member{}
	}
	return result, nil
}

func (m *mockMemberRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	return member, nil
}

func (m *mockMemberRepo) GetByIDPublic(ctx context.Context, id uuid.UUID) (*model.PublicMember, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	if !member.IsActive {
		return nil, nil
	}
	return &model.PublicMember{
		ID:   member.ID,
		Name: member.Name,
	}, nil
}

func (m *mockMemberRepo) GetByEmail(ctx context.Context, email string) (*model.Member, error) {
	member, ok := m.byEmail[email]
	if !ok {
		return nil, nil
	}
	return member, nil
}

func (m *mockMemberRepo) Update(ctx context.Context, id uuid.UUID, req model.UpdateMemberRequest) (*model.Member, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
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
	if req.IsAdmin != nil {
		member.IsAdmin = *req.IsAdmin
	}
	if req.IsActive != nil {
		member.IsActive = *req.IsActive
	}
	return member, nil
}

func (m *mockMemberRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.members, id)
	return nil
}

func (m *mockMemberRepo) HasRSVPs(ctx context.Context, memberID uuid.UUID) (bool, error) {
	return m.rsvpCounts[memberID] > 0, nil
}

func (m *mockMemberRepo) DistinctSkills(ctx context.Context) ([]string, error) {
	seen := map[string]bool{}
	var skills []string
	for _, member := range m.members {
		if member.IsActive && member.Skills != nil {
			for _, s := range member.Skills {
				if !seen[s] {
					seen[s] = true
					skills = append(skills, s)
				}
			}
		}
	}
	if skills == nil {
		skills = []string{}
	}
	return skills, nil
}

// mockEmailSender tracks invitation calls.
type mockEmailSender struct {
	calls []string
}

func (m *mockEmailSender) SendInvitation(ctx context.Context, toEmail, toName, frontendURL string) {
	m.calls = append(m.calls, toEmail)
}

func TestCreateMember_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	member, err := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "jane@example.com",
		Name:  "Jane Doe",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", member.Email)
	}
	if member.Name != "Jane Doe" {
		t.Errorf("expected name Jane Doe, got %s", member.Name)
	}
	if !member.IsActive {
		t.Error("expected new member to be active")
	}
}

func TestCreateMember_ValidationErrors(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	tests := []struct {
		name    string
		req     model.CreateMemberRequest
		details map[string]string
	}{
		{
			name: "missing email and name",
			req:  model.CreateMemberRequest{},
			details: map[string]string{
				"email": "required",
				"name":  "required",
			},
		},
		{
			name: "invalid email",
			req: model.CreateMemberRequest{
				Email: "not-an-email",
				Name:  "Jane",
			},
			details: map[string]string{
				"email": "invalid email format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tt.req)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			for field, msg := range tt.details {
				if ve.Details[field] != msg {
					t.Errorf("expected detail %s=%s, got %s", field, msg, ve.Details[field])
				}
			}
		})
	}
}

func TestCreateMember_DuplicateEmail(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	_, err := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "jane@example.com",
		Name:  "Jane Doe",
	})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "jane@example.com",
		Name:  "Jane Doe 2",
	})
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestCreateMember_StripsTelegramAt(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	handle := "@janedoe"
	member, err := svc.Create(context.Background(), model.CreateMemberRequest{
		Email:          "jane@example.com",
		Name:           "Jane Doe",
		TelegramHandle: &handle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.TelegramHandle == nil || *member.TelegramHandle != "janedoe" {
		t.Errorf("expected telegram handle 'janedoe', got %v", member.TelegramHandle)
	}
}

func TestCreateMember_EmailNormalization(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	member, err := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "  JANE@Example.COM  ",
		Name:  "  Jane Doe  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.Email != "jane@example.com" {
		t.Errorf("expected normalized email, got %s", member.Email)
	}
	if member.Name != "Jane Doe" {
		t.Errorf("expected trimmed name, got %s", member.Name)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, nil, "")

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdate_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	member, _ := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "jane@example.com",
		Name:  "Jane Doe",
	})

	newName := "Jane Smith"
	updated, err := svc.Update(context.Background(), member.ID, model.UpdateMemberRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Jane Smith" {
		t.Errorf("expected name Jane Smith, got %s", updated.Name)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, nil, "")

	name := "Jane"
	_, err := svc.Update(context.Background(), uuid.New(), model.UpdateMemberRequest{Name: &name})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdate_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, nil, "")

	emptyName := "  "
	_, err := svc.Update(context.Background(), uuid.New(), model.UpdateMemberRequest{Name: &emptyName})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["name"] != "cannot be empty" {
		t.Errorf("expected 'cannot be empty', got %s", ve.Details["name"])
	}
}

func TestUpdate_StripsTelegramAt(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	member, _ := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "jane@example.com",
		Name:  "Jane Doe",
	})

	handle := "@newhandle"
	updated, err := svc.Update(context.Background(), member.ID, model.UpdateMemberRequest{
		TelegramHandle: &handle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.TelegramHandle == nil || *updated.TelegramHandle != "newhandle" {
		t.Errorf("expected telegram handle 'newhandle', got %v", updated.TelegramHandle)
	}
}

func TestDelete_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	member, _ := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "jane@example.com",
		Name:  "Jane Doe",
	})

	err := svc.Delete(context.Background(), member.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, nil, "")

	err := svc.Delete(context.Background(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_HasRSVPs(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	member, _ := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "jane@example.com",
		Name:  "Jane Doe",
	})

	repo.rsvpCounts[member.ID] = 3

	err := svc.Delete(context.Background(), member.ID)
	if !errors.Is(err, ErrHasRSVPs) {
		t.Fatalf("expected ErrHasRSVPs, got %v", err)
	}
}

func TestList_ActiveFilter(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	// Create two members
	svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "active@example.com",
		Name:  "Active",
	})
	m2, _ := svc.Create(context.Background(), model.CreateMemberRequest{
		Email: "inactive@example.com",
		Name:  "Inactive",
	})

	// Deactivate one
	isActive := false
	svc.Update(context.Background(), m2.ID, model.UpdateMemberRequest{IsActive: &isActive})

	// List active only
	active, _ := svc.List(context.Background(), "true")
	if len(active) != 1 {
		t.Errorf("expected 1 active member, got %d", len(active))
	}

	// List all
	all, _ := svc.List(context.Background(), "all")
	if len(all) != 2 {
		t.Errorf("expected 2 total members, got %d", len(all))
	}

	// List inactive
	inactive, _ := svc.List(context.Background(), "false")
	if len(inactive) != 1 {
		t.Errorf("expected 1 inactive member, got %d", len(inactive))
	}
}

func TestCreateMember_EmptyTelegramHandleBecomesNil(t *testing.T) {
	repo := newMockRepo()
	svc := NewMemberService(repo, &mockEmailSender{}, "http://localhost:5173")

	handle := "@"
	member, err := svc.Create(context.Background(), model.CreateMemberRequest{
		Email:          "jane@example.com",
		Name:           "Jane Doe",
		TelegramHandle: &handle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.TelegramHandle != nil {
		t.Errorf("expected nil telegram handle, got %v", *member.TelegramHandle)
	}
}
