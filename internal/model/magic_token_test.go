package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMagicTokenJSONOmitsHash(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	token := MagicToken{
		ID:        uuid.New(),
		MemberID:  uuid.New(),
		TokenHash: "abc123secrethash",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}

	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := raw["token_hash"]; ok {
		t.Error("token_hash should be omitted from JSON (json:\"-\")")
	}
}

func TestMagicTokenUsedAtNullable(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	// Unused token — used_at should be null
	unused := MagicToken{
		ID:        uuid.New(),
		MemberID:  uuid.New(),
		TokenHash: "hash1",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}
	data, _ := json.Marshal(unused)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if raw["used_at"] != nil {
		t.Errorf("unused token used_at should be null, got %v", raw["used_at"])
	}

	// Used token — used_at should have a value
	usedAt := now
	used := MagicToken{
		ID:        uuid.New(),
		MemberID:  uuid.New(),
		TokenHash: "hash2",
		ExpiresAt: now.Add(15 * time.Minute),
		UsedAt:    &usedAt,
		CreatedAt: now,
	}
	data, _ = json.Marshal(used)
	json.Unmarshal(data, &raw)
	if raw["used_at"] == nil {
		t.Error("used token used_at should not be null")
	}
}
