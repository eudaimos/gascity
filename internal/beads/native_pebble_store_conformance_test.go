package beads_test

import (
	"testing"

	"github.com/gastownhall/gascity/internal/beads"
	"github.com/gastownhall/gascity/internal/beads/beadstest"
)

// TestNativePebbleStoreConformance runs the store conformance suites the
// native Dolt store runs, over the same NativeDoltStore wrapper on the beads
// library's pebble backend instead of the in-memory fixture: the first run of
// these suites against a real beads storage engine rather than a stand-in.
func TestNativePebbleStoreConformance(t *testing.T) {
	beadstest.RunStoreTests(t, func() beads.Store { return beads.NewNativePebbleStoreForConformance(t) })
	beadstest.RunCloseReasonTests(t, func() beads.Store { return beads.NewNativePebbleStoreForConformance(t) })
	beadstest.RunCloseNotificationTests(t, func(t *testing.T) beads.Store { return beads.NewNativePebbleStoreForConformance(t) })
	beadstest.RunPinnedIDDuplicateConformance(t, func(t *testing.T) beads.Store {
		return beads.NewNativePebbleStoreForPinnedIDFenceConformance(t, "kitn")
	})
}

// TestNativePebbleStoreMetadataConformance runs the metadata suite over the
// pebble backend.
func TestNativePebbleStoreMetadataConformance(t *testing.T) {
	beadstest.RunMetadataTests(t, func() beads.Store { return beads.NewNativePebbleStoreForConformance(t) })
}

// TestNativePebbleStorePinnedIDFenceConformance runs the fenced-Create suite
// over the pebble backend.
func TestNativePebbleStorePinnedIDFenceConformance(t *testing.T) {
	beadstest.RunPinnedIDFenceConformance(t, func(t *testing.T, mintPrefix string, namespaces ...string) beads.Store {
		t.Helper()
		return beads.NewNativePebbleStoreForPinnedIDFenceConformance(t, mintPrefix, namespaces...)
	})
}
