package application

import (
	"regexp"
	"strings"
	"testing"
)

func assertQueryHasPattern(t *testing.T, query, pattern string) {
	t.Helper()
	re := regexp.MustCompile(pattern)
	if !re.MatchString(query) {
		t.Fatalf("expected query to match pattern %q, got: %s", pattern, query)
	}
}

func assertArgsContain(t *testing.T, args []interface{}, want ...string) {
	t.Helper()
	for _, w := range want {
		found := false
		for _, a := range args {
			if s, ok := a.(string); ok && s == w {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected args to contain %q, got: %#v", w, args)
		}
	}
}

func TestBuildFilteredQuery_AppliesScalarAndDateFilters(t *testing.T) {
	base := `SELECT id::text, membership_status, joined_at FROM members`
	filters := map[string]string{
		"location_id": "loc-1",
		"status":      "active",
		"from":        "2026-01-01",
		"to":          "2026-01-31",
		"ignored":     "unused",
	}

	q, args := buildFilteredQuery(
		base,
		filters,
		map[string]string{
			"location_id": "location_id",
			"status":      "membership_status",
			"_date":       "joined_at",
		},
		" ORDER BY joined_at DESC LIMIT 500",
	)

	if !strings.HasPrefix(q, base) {
		t.Fatalf("expected query to start with base query, got: %s", q)
	}
	assertQueryHasPattern(t, q, `location_id = \$\d+`)
	assertQueryHasPattern(t, q, `membership_status = \$\d+`)
	assertQueryHasPattern(t, q, `joined_at >= \$\d+`)
	assertQueryHasPattern(t, q, `joined_at <= \$\d+`)
	if !strings.HasSuffix(q, " ORDER BY joined_at DESC LIMIT 500") {
		t.Fatalf("expected query suffix to be preserved, got: %s", q)
	}

	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d (%#v)", len(args), args)
	}
	assertArgsContain(t, args, "loc-1", "active", "2026-01-01", "2026-01-31")
}

func TestBuildFilteredQuery_AppendsToExistingWhereClause(t *testing.T) {
	base := `SELECT id::text FROM members WHERE membership_status IN ('expired','cancelled')`
	filters := map[string]string{
		"location_id": "loc-2",
		"from":        "2026-02-01",
		"to":          "2026-02-28",
	}

	q, args := buildFilteredQuery(
		base,
		filters,
		map[string]string{
			"location_id": "location_id",
			"_date":       "updated_at",
		},
		" ORDER BY updated_at DESC LIMIT 500",
	)

	if strings.Count(strings.ToUpper(q), " WHERE ") != 1 {
		t.Fatalf("expected a single WHERE clause, got: %s", q)
	}
	assertQueryHasPattern(t, q, `AND location_id = \$\d+`)
	assertQueryHasPattern(t, q, `updated_at >= \$\d+`)
	assertQueryHasPattern(t, q, `updated_at <= \$\d+`)

	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d (%#v)", len(args), args)
	}
	assertArgsContain(t, args, "loc-2", "2026-02-01", "2026-02-28")
}

func TestBuildFilteredQuery_ReportSpecificMappings(t *testing.T) {
	t.Run("coach_productivity", func(t *testing.T) {
		q, args := buildFilteredQuery(
			`SELECT c.id::text FROM coaches c LEFT JOIN members m ON m.location_id = c.location_id`,
			map[string]string{
				"location_id": "loc-1",
				"coach_id":    "coach-7",
			},
			map[string]string{
				"location_id": "c.location_id",
				"coach_id":    "c.id",
			},
			" GROUP BY c.id",
		)

		assertQueryHasPattern(t, q, `c\.location_id = \$\d+`)
		assertQueryHasPattern(t, q, `c\.id = \$\d+`)
		if len(args) != 2 {
			t.Fatalf("expected 2 args, got %d (%#v)", len(args), args)
		}
		assertArgsContain(t, args, "loc-1", "coach-7")
	})

	t.Run("landed_cost_report", func(t *testing.T) {
		q, args := buildFilteredQuery(
			`SELECT item_id::text, purchase_order_id::text, period FROM landed_cost_entries`,
			map[string]string{
				"item_id":           "item-1",
				"purchase_order_id": "po-9",
				"period":            "2026-04",
				"from":              "2026-04-01",
				"to":                "2026-04-30",
			},
			map[string]string{
				"item_id":           "item_id",
				"purchase_order_id": "purchase_order_id",
				"period":            "period",
				"_date":             "created_at",
			},
			" ORDER BY created_at DESC LIMIT 500",
		)

		assertQueryHasPattern(t, q, `item_id = \$\d+`)
		assertQueryHasPattern(t, q, `purchase_order_id = \$\d+`)
		assertQueryHasPattern(t, q, `period = \$\d+`)
		assertQueryHasPattern(t, q, `created_at >= \$\d+`)
		assertQueryHasPattern(t, q, `created_at <= \$\d+`)
		if len(args) != 5 {
			t.Fatalf("expected 5 args, got %d (%#v)", len(args), args)
		}
		assertArgsContain(t, args, "item-1", "po-9", "2026-04", "2026-04-01", "2026-04-30")
	})
}
