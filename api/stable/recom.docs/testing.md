# Testing Recommendations

Missing test coverage, test infrastructure improvements, QA practices.

---

## 1. Missing Bounding-Box Tests for Fasilitas 🟠 HIGH

**Location:** `integration/fasilitas_test.go`

**Issue:**

Fasilitas supports bounding box queries but they are NOT tested in integration suite. The `List(ListFasilitasInput{BBox})` flow is critical for map UI but unverified.

**Impact:**
- Bounding box query may have bugs that only surface in production
- Map-based UIs may show wrong results

**Recommendation:**

Add tests:
```go
func TestFasilitas_BoundingBox(t *testing.T) {
    // Setup: create 4 fasilitas at known coordinates
    f1 := createFasilitas(t, -6.1, 106.1)  // inside
    f2 := createFasilitas(t, -6.2, 106.2)  // inside
    f3 := createFasilitas(t, -7.0, 107.0)  // outside (south/east)
    f4 := createFasilitas(t, -5.0, 105.0)  // outside (north/west)
    
    tests := []struct {
        name  string
        bbox  BoundingBox
        want  []string  // expected IDs
    }{
        {
            name: "exact bbox",
            bbox: BoundingBox{MinLat: -6.3, MaxLat: -6.05, MinLon: 106.05, MaxLon: 106.25},
            want: []string{f1.ID, f2.ID},
        },
        {
            name: "inverted bbox",
            bbox: BoundingBox{MinLat: 106.1, MaxLat: 106.05, MinLon: -6.1, MaxLon: -6.05},
            wantErr: true,  // validation should reject
        },
        // ... more cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Make request, validate
        })
    }
}
```

**Effort:** S (1-2 days)

---

## 2. No Backup Integration Tests 🟡 MEDIUM

**Location:** `integration/backup*.go` (doesn't exist)

**Issue:**

Backup feature is completely untested. The only verification is manual CLI execution.

**Impact:**
- pg_dump/pg_restore failures only caught in production
- PGPASSWORD injection not verified
- File permission/security not verified

**Recommendation:**

If HTTP backup routes get enabled ([bugs.md](./bugs.md#4)), add tests:
```go
func TestBackup_Create(t *testing.T) {
    // POST /api/v1/backups
    // Verify: file exists on disk
    // Verify: row in backups table
}

func TestBackup_Restore(t *testing.T) {
    // Setup: create test data
    // Create backup
    // TRUNCATE all tables
    // Restore
    // Verify: data back
}

func TestBackup_Download(t *testing.T) {
    // GET /api/v1/backups/{id}/download
    // Verify: returns binary with correct headers
}
```

**Effort:** M (3-5 days)

---

## 3. No Email Notification Tests for PPID 🟡 MEDIUM

**Location:** `integration/ppid_approval_test.go`

**Issue:**

The approval workflow sends email then updates DB but tests don't verify email was sent (e.g., via test SMTP catcher).

**Recommendation:**

Use a test SMTP server:
```go
import "github.com/mailhog/MailHog"  // or similar

func TestPPID_ApproveSendsEmail(t *testing.T) {
    // Set up mail catcher
    smtp := startMailCatcher(t)
    defer smtp.Close()
    
    cfg.SMTP.Host = smtp.Host
    cfg.SMTP.Port = smtp.Port
    
    // ... approve request ...
    
    // Verify email received
    msg := smtp.LastMessage()
    require.Contains(t, msg.Body, "Download Document")
    require.Equal(t, requesterEmail, msg.To[0])
}
```

**Effort:** M (2-3 days)

---

## 4. No Chaos/Failure Tests 🟡 MEDIUM

**Location:** `integration/`

**Issue:**

Tests assume happy path. No tests for:
- DB connection drops
- SMTP timeout
- File upload failure
- Concurrent requests

**Recommendation:**

Add chaos tests:
```go
func TestBerita_ConcurrentUpdate(t *testing.T) {
    // Create berita
    // Spawn 10 goroutines updating concurrently
    // Verify: no race, no panic, consistent final state
}

func TestAuth_DBDown(t *testing.T) {
    // Stop postgres container
    // Try to login
    // Verify: 500 error, not crash
}
```

**Effort:** L (1-2 weeks)

---

## 5. No Property-Based Tests 🟡 MEDIUM

**Location:** `domain/`, validators

**Issue:**

Validation logic tested with a few cases. Property-based testing (with `testing/quick` or `gopter`) would find edge cases.

**Recommendation:**

For complex validators:
```go
func TestUser_Validate_Properties(t *testing.T) {
    f := func(name, email, password string) bool {
        u := &User{Name: name, Email: email, HashedPassword: password}
        err := u.Validate()
        if err == nil {
            // Validate successful cases
            return len(name) <= 255 && strings.Contains(email, "@")
        }
        // Validate failure cases
        return true
    }
    
    if err := quick.Check(f, nil); err != nil {
        t.Fatal(err)
    }
}
```

**Effort:** S (per domain, 1 day each)

---

## 6. No Performance/Benchmark Tests 🟡 MEDIUM

**Location:** `*_test.go` (Benchmark functions)

**Issue:**

Zero benchmarks. No way to detect performance regressions.

**Recommendation:**

Add benchmark functions:
```go
func BenchmarkBeritaList(b *testing.B) {
    setup(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.List(ctx, ListBeritaInput{Page: 1, Limit: 20})
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

Run with `go test -bench=. -benchmem`.

**Effort:** S (per feature, few hours)

---

## 7. No Coverage Threshold Enforcement 🟡 MEDIUM

**Location:** Makefile, CI

**Issue:**

Coverage files exist (`coverage.out`) but no minimum threshold is enforced. Coverage can silently drop.

**Recommendation:**

```makefile
coverage-threshold: ## Verify coverage >= 80%
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | awk '/total:/ {gsub("%","",$$3); if($$3 < 80) {print "FAIL: coverage "$$3"% < 80%"; exit 1}}'
```

Or in CI:
```yaml
- name: Check coverage
  run: |
    COVERAGE=$(go test ./... -cover | grep -oP '\d+(?=\.\d+%)' | head -1)
    if [ "$COVERAGE" -lt 80 ]; then exit 1; fi
```

**Effort:** XS (1 hour)

---

## 8. No Mutation Testing 🟢 LOW

**Location:** CI

**Issue:**

Tests can pass even if implementation is wrong (e.g., assert always true). Mutation testing would catch this.

**Recommendation:**

Use `go-mutesting` or `gremlins` in CI to verify tests actually catch bugs.

**Effort:** M (setup + CI integration, ~2 days)

---

## 9. No E2E Tests 🟢 LOW

**Location:** `e2e/` (doesn't exist)

**Issue:**

Integration tests cover API endpoints, but there's no end-to-end test that simulates a full user journey:
1. Login as admin
2. Create berita with image
3. Verify appears in public list

**Recommendation:**

Add e2e test suite using a test framework like `ginkgo` or just plain `go test` with scenarios.

**Effort:** L (1-2 weeks)

---

## Summary Table

| # | Gap | Priority | Effort |
|---|---|---|---|
| 1 | Bounding-box tests | 🟠 HIGH | S |
| 2 | Backup integration tests | 🟡 MEDIUM | M |
| 3 | Email notification tests | 🟡 MEDIUM | M |
| 4 | Chaos tests | 🟡 MEDIUM | L |
| 5 | Property-based tests | 🟡 MEDIUM | S |
| 6 | Benchmark tests | 🟡 MEDIUM | S |
| 7 | Coverage threshold | 🟡 MEDIUM | XS |
| 8 | Mutation testing | 🟢 LOW | M |
| 9 | E2E tests | 🟢 LOW | L |