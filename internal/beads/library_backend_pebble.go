package beads

// The pebble beads backend is a pure-Go key-value store that lives inside
// a workspace's .beads directory and is opened through the linked beads
// library alone: no bd process, no Dolt server, no SQL. Importing the public
// registrant package puts it in the library's backend registry, which is what
// contract.IsLibraryExtensionBackend requires beside gc's own allowlist for a
// scope whose metadata.json names it to open through the native store.
import _ "github.com/steveyegge/beads/backend/pebble" // registers the pebble backend
