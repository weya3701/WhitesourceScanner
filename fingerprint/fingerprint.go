package fingerprint

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
)

const (
	typeFile    byte = 1
	typeDir     byte = 2
	typeSymlink byte = 3
)

var (
	dirDomain     = []byte("filesystem-fingerprint\x00v1\x00directory\x00")
	symlinkDomain = []byte("filesystem-fingerprint\x00v1\x00symlink\x00")
)

type scanner struct {
	root    string
	rules   []ignoreRule
	entries []Entry
}

type child struct {
	name   string
	kind   byte
	digest [sha256.Size]byte
}

// Hash returns only the deterministic SHA-256 fingerprint of a file,
// directory tree, or symbolic link. It applies the same ignore options as
// Scan and does not collect manifest entries.
func Hash(path string, options Options) (string, error) {
	options.IncludeEntries = false
	manifest, err := Scan(path, options)
	if err != nil {
		return "", err
	}
	return manifest.RootHash, nil
}

func Scan(path string, options Options) (Manifest, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("resolve root: %w", err)
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return Manifest{}, fmt.Errorf("inspect root: %w", err)
	}

	patterns := make([]string, 0)
	if options.UseDefaults {
		patterns = append(patterns, defaultIgnorePatterns...)
	}
	if options.IgnoreFile != "" {
		loaded, err := readIgnoreFile(options.IgnoreFile)
		if err != nil {
			return Manifest{}, err
		}
		patterns = append(patterns, loaded...)
	}
	patterns = append(patterns, options.IgnorePatterns...)
	rules, err := compileIgnoreRules(patterns)
	if err != nil {
		return Manifest{}, err
	}

	s := scanner{root: absolute, rules: rules}
	digest, rootType, err := s.hashNode(absolute, "", info, true)
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{
		Format:    Format,
		Algorithm: "sha256",
		RootHash:  hex.EncodeToString(digest[:]),
		RootType:  rootType,
		Generated: time.Now().UTC(),
		Ignore:    patterns,
	}
	if options.IncludeEntries {
		sort.Slice(s.entries, func(i, j int) bool {
			return bytes.Compare([]byte(s.entries[i].Path), []byte(s.entries[j].Path)) < 0
		})
		manifest.Entries = s.entries
	}
	return manifest, nil
}

func (s *scanner) hashNode(fullPath, relative string, info fs.FileInfo, root bool) ([sha256.Size]byte, string, error) {
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(fullPath)
		if err != nil {
			return [sha256.Size]byte{}, "", fmt.Errorf("read symlink %q: %w", displayPath(relative), err)
		}
		normalizedTarget := normalizePath(target)
		digest := hashLengthPrefixed(symlinkDomain, []byte(normalizedTarget))
		s.addEntry(relative, "symlink", digest, "", normalizedTarget, info, root)
		return digest, "symlink", nil
	case info.IsDir():
		digest, err := s.hashDirectory(fullPath, relative)
		if err != nil {
			return [sha256.Size]byte{}, "", err
		}
		s.addEntry(relative, "directory", digest, "", "", info, root)
		return digest, "directory", nil
	case info.Mode().IsRegular():
		digest, size, err := hashFile(fullPath)
		if err != nil {
			return [sha256.Size]byte{}, "", fmt.Errorf("hash file %q: %w", displayPath(relative), err)
		}
		s.addFileEntry(relative, digest, size, info, root)
		return digest, "file", nil
	default:
		return [sha256.Size]byte{}, "", fmt.Errorf("unsupported filesystem object %q (%s)", displayPath(relative), info.Mode())
	}
}

func (s *scanner) hashDirectory(fullPath, relative string) ([sha256.Size]byte, error) {
	items, err := os.ReadDir(fullPath)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("read directory %q: %w", displayPath(relative), err)
	}
	children := make([]child, 0, len(items))
	seen := make(map[string]string)
	for _, item := range items {
		name := norm.NFC.String(item.Name())
		if prior, exists := seen[name]; exists {
			return [sha256.Size]byte{}, fmt.Errorf("Unicode-normalized name collision in %q: %q and %q", displayPath(relative), prior, item.Name())
		}
		seen[name] = item.Name()
		childRelative := name
		if relative != "" {
			childRelative = relative + "/" + name
		}
		info, err := os.Lstat(filepath.Join(fullPath, item.Name()))
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("inspect %q: %w", childRelative, err)
		}
		if ignored(childRelative, info.IsDir(), s.rules) {
			continue
		}
		digest, kindName, err := s.hashNode(filepath.Join(fullPath, item.Name()), childRelative, info, false)
		if err != nil {
			return [sha256.Size]byte{}, err
		}
		children = append(children, child{name: name, kind: kindByte(kindName), digest: digest})
	}
	sort.Slice(children, func(i, j int) bool {
		return bytes.Compare([]byte(children[i].name), []byte(children[j].name)) < 0
	})
	h := sha256.New()
	_, _ = h.Write(dirDomain)
	for _, item := range children {
		_, _ = h.Write([]byte{item.kind})
		writeLengthPrefixed(h, []byte(item.name))
		_, _ = h.Write(item.digest[:])
	}
	var result [sha256.Size]byte
	copy(result[:], h.Sum(nil))
	return result, nil
}

func hashFile(path string) ([sha256.Size]byte, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return [sha256.Size]byte{}, 0, err
	}
	defer file.Close()
	h := sha256.New()
	size, err := io.Copy(h, file)
	if err != nil {
		return [sha256.Size]byte{}, 0, err
	}
	var result [sha256.Size]byte
	copy(result[:], h.Sum(nil))
	return result, size, nil
}

func (s *scanner) addFileEntry(path string, digest [sha256.Size]byte, size int64, info fs.FileInfo, root bool) {
	if root && path == "" {
		path = "."
	}
	s.entries = append(s.entries, Entry{
		Path:        path,
		Type:        "file",
		ContentHash: hex.EncodeToString(digest[:]),
		Size:        size,
		Metadata:    metadata(info),
	})
}

func (s *scanner) addEntry(path, kind string, digest [sha256.Size]byte, contentHash, target string, info fs.FileInfo, root bool) {
	if root && path == "" {
		path = "."
	}
	entry := Entry{Path: path, Type: kind, Metadata: metadata(info), LinkTarget: target}
	if kind == "directory" {
		entry.TreeHash = hex.EncodeToString(digest[:])
	}
	s.entries = append(s.entries, entry)
}

func metadata(info fs.FileInfo) Metadata {
	return Metadata{Mode: info.Mode().String(), ModifiedAt: info.ModTime().UTC()}
}

func readIgnoreFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open ignore file: %w", err)
	}
	defer file.Close()
	var patterns []string
	input := bufio.NewScanner(file)
	for input.Scan() {
		patterns = append(patterns, input.Text())
	}
	if err := input.Err(); err != nil {
		return nil, fmt.Errorf("read ignore file: %w", err)
	}
	return patterns, nil
}

func normalizePath(value string) string {
	return norm.NFC.String(strings.ReplaceAll(value, `\`, "/"))
}

func hashLengthPrefixed(domain, value []byte) [sha256.Size]byte {
	h := sha256.New()
	_, _ = h.Write(domain)
	writeLengthPrefixed(h, value)
	var result [sha256.Size]byte
	copy(result[:], h.Sum(nil))
	return result
}

func writeLengthPrefixed(w io.Writer, value []byte) {
	_ = binary.Write(w, binary.BigEndian, uint32(len(value)))
	_, _ = w.Write(value)
}

func kindByte(kind string) byte {
	switch kind {
	case "file":
		return typeFile
	case "directory":
		return typeDir
	default:
		return typeSymlink
	}
}

func displayPath(path string) string {
	if path == "" {
		return "."
	}
	return path
}

func Verify(manifest Manifest, path string, options Options) (Verification, error) {
	if manifest.Format != Format {
		return Verification{}, fmt.Errorf("unsupported manifest format %q", manifest.Format)
	}
	if manifest.Algorithm != "sha256" {
		return Verification{}, fmt.Errorf("unsupported algorithm %q", manifest.Algorithm)
	}
	options.IgnorePatterns = append([]string(nil), manifest.Ignore...)
	options.IgnoreFile = ""
	options.UseDefaults = false
	options.IncludeEntries = false
	actual, err := Scan(path, options)
	if err != nil {
		return Verification{}, err
	}
	if len(manifest.RootHash) != sha256.Size*2 {
		return Verification{}, errors.New("manifest rootHash is not a SHA-256 digest")
	}
	return Verification{
		Match:    strings.EqualFold(manifest.RootHash, actual.RootHash),
		Expected: manifest.RootHash,
		Actual:   actual.RootHash,
	}, nil
}
