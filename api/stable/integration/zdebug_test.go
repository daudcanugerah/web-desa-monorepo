package integration

import (
	"context"
	"fmt"
	"testing"
)

func TestZDebugCleanDB(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)
	ctx := context.Background()

	// Insert a sentinel user BEFORE CleanDatabase would run.
	_, _ = ts.DB.ExecContext(ctx, `INSERT INTO users (id, name, email, hashed_password, created_at, updated_at)
		VALUES ('00000000-0000-0000-0000-000000000099', 'Sentinel', 'sentinel@test.com', 'x', NOW(), NOW())`)

	rows, _ := ts.DB.QueryContext(ctx, `SELECT email FROM users`)
	defer rows.Close()
	for rows.Next() {
		var e string
		_ = rows.Scan(&e)
		fmt.Printf("user: %s\n", e)
	}
}

func TestZDebugAfterClean(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)
	ctx := context.Background()
	rows, _ := ts.DB.QueryContext(ctx, `SELECT email FROM users`)
	defer rows.Close()
	fmt.Println("=== users after CleanDatabase ===")
	for rows.Next() {
		var e string
		_ = rows.Scan(&e)
		fmt.Printf("user: %s\n", e)
	}
}