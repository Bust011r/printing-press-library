// Copyright 2026 bust011r and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mvanhorn/printing-press-library/library/travel/hostex/internal/store"
)

type fixedPageClient struct{ body string }

func (c *fixedPageClient) Get(_ context.Context, _ string, _ map[string]string) (json.RawMessage, error) {
	return json.RawMessage(c.body), nil
}

func (c *fixedPageClient) RateLimit() float64 { return 0 }

// PATCH(hostex-sync-reconcile-skips-partial-windows): a full sync prunes only
// when the fetch enumerated the whole table. The transactions window, --since
// and --param filters are partial, so rows outside them must survive.
func TestFullSyncPrunesOnlyWholeTableFetches(t *testing.T) {
	cases := []struct {
		name      string
		resource  string
		body      string
		since     string
		wantAfter int
	}{
		{"transactions window keeps older rows", "transactions", `{"error_code":200,"data":{"transactions":[{"id":"new-1","amount":1}]}}`, "", 2},
		{"since keeps older rows", "conversations", `{"error_code":200,"data":{"conversations":[{"id":"new-1"}]}}`, "2026-01-01T00:00:00Z", 2},
		{"whole-table fetch prunes rows the API dropped", "conversations", `{"error_code":200,"data":{"conversations":[{"id":"new-1"}]}}`, "", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := store.Open(filepath.Join(t.TempDir(), "data.db"))
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer db.Close()
			if err := db.Upsert(tc.resource, "old-1", json.RawMessage(`{"id":"old-1"}`)); err != nil {
				t.Fatalf("seed: %v", err)
			}
			syncResource(context.Background(), &fixedPageClient{body: tc.body}, db, tc.resource, tc.since, true, 1, false, true, nil, nil)
			rows, err := db.List(tc.resource, 10)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(rows) != tc.wantAfter {
				t.Fatalf("%s rows after full sync = %d, want %d", tc.resource, len(rows), tc.wantAfter)
			}
		})
	}
}
