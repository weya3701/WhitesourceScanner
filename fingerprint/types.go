package fingerprint

import "time"

const Format = "filesystem-fingerprint-v1"

type Options struct {
	IgnorePatterns []string
	IgnoreFile     string
	UseDefaults    bool
	IncludeEntries bool
}

type Manifest struct {
	Format    string    `json:"format"`
	Algorithm string    `json:"algorithm"`
	RootHash  string    `json:"rootHash"`
	RootType  string    `json:"rootType"`
	Generated time.Time `json:"generatedAt"`
	Ignore    []string  `json:"ignorePatterns,omitempty"`
	Entries   []Entry   `json:"entries,omitempty"`
}

type Entry struct {
	Path        string   `json:"path"`
	Type        string   `json:"type"`
	ContentHash string   `json:"contentHash,omitempty"`
	TreeHash    string   `json:"treeHash,omitempty"`
	LinkTarget  string   `json:"linkTarget,omitempty"`
	Size        int64    `json:"size,omitempty"`
	Metadata    Metadata `json:"metadata"`
}

// Metadata is informational and never contributes to RootHash.
type Metadata struct {
	Mode       string    `json:"mode"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type Verification struct {
	Match    bool   `json:"match"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}
