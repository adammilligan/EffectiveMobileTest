package subscriptions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/migrator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestRepo_Integration_CRUD(t *testing.T) {
	ctx := context.Background()

	pg, dsn := startPostgres(t, ctx)
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	t.Cleanup(pool.Close)

	sqlDB := stdlib.OpenDBFromPool(pool)
	if err := migrateForTest(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewRepo(pool)

	start, _ := ParseMonth("07-2025")
	end, _ := ParseMonth("09-2025")

	created, err := repo.Create(ctx, CreateParams{
		ServiceName: "Yandex Plus",
		PriceRub:    400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartMonth:  start,
		EndMonth:    &end,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if created.ID == "" {
		t.Fatalf("expected id")
	}

	got, isFound, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if !isFound {
		t.Fatalf("expected found")
	}

	if got.ServiceName != "Yandex Plus" || got.PriceRub != 400 {
		t.Fatalf("unexpected subscription: %+v", got)
	}

	list, err := repo.List(ctx, ListParams{UserID: &created.UserID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) == 0 {
		t.Fatalf("expected non-empty list")
	}

	newName := "Yandex Plus 2"
	newPrice := 500

	updated, isFound, err := repo.Update(ctx, created.ID, UpdateParams{
		ServiceName: &newName,
		PriceRub:    &newPrice,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if !isFound {
		t.Fatalf("expected found on update")
	}

	if updated.ServiceName != newName || updated.PriceRub != newPrice {
		t.Fatalf("unexpected updated: %+v", updated)
	}

	isDeleted, err := repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	if !isDeleted {
		t.Fatalf("expected deleted")
	}
}

func TestRepo_Integration_TotalQuerySelection(t *testing.T) {
	ctx := context.Background()

	pg, dsn := startPostgres(t, ctx)
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	t.Cleanup(pool.Close)

	sqlDB := stdlib.OpenDBFromPool(pool)
	if err := migrateForTest(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewRepo(pool)

	userID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"
	jul, _ := ParseMonth("07-2025")
	sep, _ := ParseMonth("09-2025")
	dec, _ := ParseMonth("12-2025")

	_, err = repo.Create(ctx, CreateParams{
		ServiceName: "A",
		PriceRub:    10,
		UserID:      userID,
		StartMonth:  jul,
		EndMonth:    &sep,
	})
	if err != nil {
		t.Fatalf("create A: %v", err)
	}

	_, err = repo.Create(ctx, CreateParams{
		ServiceName: "B",
		PriceRub:    10,
		UserID:      userID,
		StartMonth:  dec,
		EndMonth:    nil,
	})
	if err != nil {
		t.Fatalf("create B: %v", err)
	}

	items, err := repo.FindOverlappingForTotal(ctx, TotalQueryParams{
		UserID:      &userID,
		ServiceName: nil,
		From:        jul,
		To:          sep,
	})
	if err != nil {
		t.Fatalf("find overlapping: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].ServiceName != "A" {
		t.Fatalf("expected A, got %s", items[0].ServiceName)
	}
}

func startPostgres(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, string) {
	t.Helper()

	pg, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("subscriptions"),
		postgres.WithUsername("subscriptions"),
		postgres.WithPassword("subscriptions"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pg.Terminate(ctx)

		t.Fatalf("dsn: %v", err)
	}

	return pg, dsn
}

func migrateForTest(db *sql.DB) error {
	return migrator.Up(db, "../../../migrations")
}

