package fingerprint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeterministicAndPathSensitive(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	writeFile(t, filepath.Join(first, "a.txt"), "same")
	writeFile(t, filepath.Join(second, "a.txt"), "same")

	a := mustScan(t, first)
	b := mustScan(t, second)
	if a.RootHash != b.RootHash {
		t.Fatalf("absolute path changed hash: %s != %s", a.RootHash, b.RootHash)
	}

	if err := os.Rename(filepath.Join(second, "a.txt"), filepath.Join(second, "b.txt")); err != nil {
		t.Fatal(err)
	}
	renamed := mustScan(t, second)
	if a.RootHash == renamed.RootHash {
		t.Fatal("renaming a file did not change root hash")
	}
	if contentHashFor(t, a, "a.txt") != contentHashFor(t, renamed, "b.txt") {
		t.Fatal("renaming a file changed its content hash")
	}
}

func TestEmptyDirectoryAndIgnoredFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	before := mustScan(t, root)
	writeFile(t, filepath.Join(root, ".DS_Store"), "ignored")
	afterIgnored := mustScan(t, root)
	if before.RootHash != afterIgnored.RootHash {
		t.Fatal("default ignored file changed root hash")
	}
	if err := os.Mkdir(filepath.Join(root, "another-empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	afterDirectory := mustScan(t, root)
	if before.RootHash == afterDirectory.RootHash {
		t.Fatal("empty directory did not change root hash")
	}
}

func TestMetadataDoesNotAffectHash(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file")
	writeFile(t, path, "content")
	before := mustScan(t, root)
	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatal(err)
	}
	after := mustScan(t, root)
	if before.RootHash != after.RootHash {
		t.Fatal("permissions changed root hash")
	}
}

func TestHashReturnsOnlyRootHash(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "file.txt"), "content")
	writeFile(t, filepath.Join(root, ".DS_Store"), "ignored")

	hash, err := Hash(root, Options{
		UseDefaults:    true,
		IncludeEntries: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	manifest := mustScan(t, root)
	if hash != manifest.RootHash {
		t.Fatalf("Hash() = %q, want %q", hash, manifest.RootHash)
	}
}

func TestSymlinkTargetIsHashedWithoutFollowing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "one"), "same")
	writeFile(t, filepath.Join(root, "two"), "same")
	link := filepath.Join(root, "link")
	if err := os.Symlink("one", link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	before := mustScan(t, root)
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("two", link); err != nil {
		t.Fatal(err)
	}
	after := mustScan(t, root)
	if before.RootHash == after.RootHash {
		t.Fatal("changing symlink target did not change root hash")
	}
}

func mustScan(t *testing.T, path string) Manifest {
	t.Helper()
	result, err := Scan(path, Options{UseDefaults: true, IncludeEntries: true})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func writeFile(t *testing.T, path, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}

func contentHashFor(t *testing.T, manifest Manifest, path string) string {
	t.Helper()
	for _, entry := range manifest.Entries {
		if entry.Path == path {
			return entry.ContentHash
		}
	}
	t.Fatalf("entry %q not found", path)
	return ""
}
