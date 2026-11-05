package beads

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	beadslib "github.com/steveyegge/beads"
	bdpebble "github.com/steveyegge/beads/backend/pebble"
)

// TestNativePebbleStoreWritesJournalOnlyWhenWorkspaceEnablesIt is
// TestNativeDoltStoreWritesJournalOnlyWhenWorkspaceEnablesIt over a pebble
// scope: gc's production open (OpenNativeDoltStoreAt) resolves the scope's
// activation and applies it to the pebble store, a native write lands in the
// journal exactly when the scope turns it on, and the store's own journal
// read (the one the ledger feed tails) serves the row or refuses as
// inactive.
func TestNativePebbleStoreWritesJournalOnlyWhenWorkspaceEnablesIt(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
		want   bool
	}{
		{name: "enabled", config: "events-journal: true\n", want: true},
		{name: "disabled by default"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("BD_EVENTS_JOURNAL", "")
			ctx := context.Background()
			scopeRoot := t.TempDir()
			beadsDir := filepath.Join(scopeRoot, ".beads")
			if err := os.MkdirAll(beadsDir, 0o755); err != nil {
				t.Fatal(err)
			}
			// What bd init --backend=pebble leaves behind: metadata.json
			// naming the backend, and the store itself under .beads.
			if err := os.WriteFile(filepath.Join(beadsDir, "metadata.json"), []byte(`{"backend":"`+bdpebble.Name+`"}`), 0o644); err != nil {
				t.Fatal(err)
			}
			seed, err := beadslib.OpenBestAvailable(ctx, beadsDir)
			if err != nil {
				t.Fatalf("OpenBestAvailable on a pebble scope: %v", err)
			}
			if err := seed.SetConfig(ctx, nativeIssuePrefixConfigKey, "gc"); err != nil {
				t.Fatalf("SetConfig(issue_prefix): %v", err)
			}
			if err := seed.Close(); err != nil {
				t.Fatalf("close seed storage: %v", err)
			}
			if tc.config != "" {
				if err := os.WriteFile(filepath.Join(beadsDir, "config.yaml"), []byte(tc.config), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			store, err := OpenNativeDoltStoreAt(ctx, scopeRoot, nil)
			if err != nil {
				t.Fatalf("OpenNativeDoltStoreAt: %v", err)
			}
			t.Cleanup(func() {
				if err := store.CloseStore(); err != nil {
					t.Fatalf("CloseStore: %v", err)
				}
			})
			if got := store.EventsJournalActive(); got != tc.want {
				t.Fatalf("EventsJournalActive() = %v, want %v", got, tc.want)
			}
			bead, err := store.Create(Bead{Title: "journal coverage"})
			if err != nil {
				t.Fatalf("Create: %v", err)
			}

			ids := nativeJournalIssueIDs(ctx, t, store.storage)
			if journaled := slices.Contains(ids, bead.ID); journaled != tc.want {
				t.Fatalf("journal recorded %s = %v, want %v; journal issue IDs: %v", bead.ID, journaled, tc.want, ids)
			}

			// The read the ledger feed tails goes through the native store,
			// which gates it on the scope's activation.
			page, err := store.ReadEventsJournalPage(ctx, 0, 0)
			if !tc.want {
				if !errors.Is(err, ErrEventsJournalInactive) {
					t.Fatalf("ReadEventsJournalPage on an inactive scope: err = %v, want ErrEventsJournalInactive", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadEventsJournalPage: %v", err)
			}
			if len(page.Rows) == 0 || page.Rows[len(page.Rows)-1].IssueID != bead.ID || page.Rows[len(page.Rows)-1].Op != "create" {
				t.Fatalf("ReadEventsJournalPage rows = %+v, want the create of %s last", page.Rows, bead.ID)
			}
			if page.Head != page.Rows[len(page.Rows)-1].Seq {
				t.Fatalf("head = %d, want the last row's seq %d", page.Head, page.Rows[len(page.Rows)-1].Seq)
			}
			journal, ok := ActiveEventsJournal(store)
			if !ok || journal == nil {
				t.Fatalf("ActiveEventsJournal(store) = %v, %v; want the store's journal", journal, ok)
			}
		})
	}
}
