package cli

import (
	"strings"
	"testing"
)

// The calendar write commands must be discoverable through `which` by a
// natural-language query, ahead of the read-side novel commands.
func TestWhichRanksCalendarWritesFirst(t *testing.T) {
	cases := map[string]string{
		"update prices of a listing":     "listings update-prices",
		"update restrictions":            "listings update-restrictions",
		"update inventories of a day":    "listings update-inventories",
		"update price":                   "listings update-prices",
		"update restriction":             "listings update-restrictions",
		"update inventory":               "listings update-inventories",
		"change the price":               "listings update-prices",
		"set a minimum stay restriction": "listings update-restrictions",
	}
	for q, want := range cases {
		got := rankWhich(whichIndex, q, 3)
		if len(got) == 0 || got[0].Entry.Command != want {
			t.Errorf("which %q: top match = %+v, want %q", q, got, want)
		}
	}
}

// A read-side query must never resolve to a write command, even when it
// shares nouns with one ("listing", "price").
func TestWhichReadQueriesDoNotRankWrites(t *testing.T) {
	cases := map[string]string{
		"check listing price drift": "price-parity",
		"price differences":         "price-parity",
		"prices across channels":    "price-parity",
		"price comparison":          "price-parity",
	}
	for q, want := range cases {
		got := rankWhich(whichIndex, q, 3)
		found := false
		for _, m := range got {
			if m.Entry.Command == want {
				found = true
			}
			if strings.Contains(m.Entry.Command, "update-") {
				t.Errorf("which %q returned write command %q", q, m.Entry.Command)
			}
		}
		if !found {
			t.Errorf("which %q: top matches = %+v, want %q among them", q, got, want)
		}
	}
}
