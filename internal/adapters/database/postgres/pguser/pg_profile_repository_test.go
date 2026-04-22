package pguser

import (
	"context"
	"os"
	"testing"
	"time"

	"vexentra-api/internal/modules/user"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newTestProfileRepo(t *testing.T) user.ProfileRepository {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("skip pg integration test: TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}

	// Keep test data isolated by using a transaction and rollback at cleanup.
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	return NewProfileRepository(tx, nil)
}

func TestCreateExperience_PersistsFields(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()
	startedAt := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	endedAt := time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)

	in := &user.Experience{
		PersonID:    uuid.New(),
		Company:     "Vexentra",
		Position:    "Backend Engineer",
		Location:    "Bangkok",
		Description: "Built API services",
		StartedAt:   startedAt,
		EndedAt:     &endedAt,
		IsCurrent:   false,
		SortOrder:   1,
	}

	if err := repo.CreateExperience(ctx, in); err != nil {
		t.Fatalf("create experience: %v", err)
	}
	if in.ID == uuid.Nil {
		t.Fatalf("expected generated ID")
	}

	got, err := repo.ListExperiencesByPersonID(ctx, in.PersonID)
	if err != nil {
		t.Fatalf("list experiences: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 experience, got %d", len(got))
	}
	if got[0].Location != "Bangkok" {
		t.Fatalf("expected location Bangkok, got %q", got[0].Location)
	}
}

func TestUpdateExperience_UpdatesLocationAndContent(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()
	personID := uuid.New()
	startedAt := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)

	in := &user.Experience{
		PersonID:    personID,
		Company:     "Old Co",
		Position:    "Developer",
		Location:    "Bangkok",
		Description: "Old description",
		StartedAt:   startedAt,
		IsCurrent:   true,
		SortOrder:   1,
	}
	if err := repo.CreateExperience(ctx, in); err != nil {
		t.Fatalf("create experience: %v", err)
	}

	endedAt := time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC)
	update := &user.Experience{
		ID:          in.ID,
		PersonID:    personID,
		Company:     "New Co",
		Position:    "Senior Developer",
		Location:    "Chiang Mai",
		Description: "Updated description",
		StartedAt:   startedAt,
		EndedAt:     &endedAt,
		IsCurrent:   false,
		SortOrder:   2,
	}
	if err := repo.UpdateExperience(ctx, update); err != nil {
		t.Fatalf("update experience: %v", err)
	}

	got, err := repo.ListExperiencesByPersonID(ctx, personID)
	if err != nil {
		t.Fatalf("list experiences: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 experience, got %d", len(got))
	}

	row := got[0]
	if row.Company != "New Co" {
		t.Fatalf("expected company New Co, got %q", row.Company)
	}
	if row.Position != "Senior Developer" {
		t.Fatalf("expected position Senior Developer, got %q", row.Position)
	}
	if row.Location != "Chiang Mai" {
		t.Fatalf("expected location Chiang Mai, got %q", row.Location)
	}
	if row.Description != "Updated description" {
		t.Fatalf("expected description updated, got %q", row.Description)
	}
	if row.EndedAt == nil || !row.EndedAt.Equal(endedAt) {
		t.Fatalf("expected ended_at %v, got %v", endedAt, row.EndedAt)
	}
}
