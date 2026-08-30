package project

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Document storage helpers (P1-005).
//
// This package stores document files on the local filesystem for the
// development MVP. The design keeps storage behind small helpers so it can be
// swapped for MinIO/S3-compatible object storage later without touching
// handlers: the metadata row keeps a RELATIVE storage key and only these
// helpers know how to translate the key into an actual read/write location.

const (
	// documentCategoryNone is the empty/unspecified category marker.
	documentCategoryNone = ""

	// maxDocumentNameLen guards display names.
	maxDocumentNameLen = 500
)

// ErrInvalidFileType is returned when an uploaded file's extension/MIME is not
// in the allowlist.
var ErrInvalidFileType = errors.New("file type is not allowed")

// ErrFileTooLarge is returned when an uploaded file exceeds the size limit.
var ErrFileTooLarge = errors.New("file exceeds maximum allowed size")

// allowedDocumentExt maps allowed extensions to their canonical MIME type.
var allowedDocumentExt = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".txt":  "text/plain",
	".csv":  "text/csv",
}

// sanitizeFilename returns a safe basename: strips directory components and
// control characters, collapses whitespace, and falls back to a generic name
// when the result would be empty.
func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	base = strings.Map(func(r rune) rune {
		if r < 32 || r == 0x7f {
			return -1
		}
		return r
	}, base)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == ".." {
		return "document"
	}
	return base
}

// validateDocumentFile checks extension allowlist + file header (magic bytes)
// and returns the canonical MIME type. Extension check happens first; when the
// header sniff disagrees with the extension, the header result is used for the
// stored MIME but the extension must still be allowed.
func validateDocumentFile(header []byte, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := allowedDocumentExt[ext]; !ok {
		return "", ErrInvalidFileType
	}

	sniffed := http.DetectContentType(header)
	if sniffed == "" {
		sniffed = allowedDocumentExt[ext]
	}
	return sniffed, nil
}

// buildDocumentStorageKey builds a RELATIVE storage key:
// documents/{orgID}/{projectID}/{documentID}/{safe_filename}.
// The key never contains absolute paths or user-controlled separators.
func buildDocumentStorageKey(orgID, projectID, documentID, safeName string) string {
	return fmt.Sprintf("documents/%s/%s/%s/%s", orgID, projectID, documentID, safeName)
}

// saveDocumentFile writes file content under the local storage root using the
// relative key. It creates parent directories as needed and refuses to write
// outside the root.
func saveDocumentFile(root, key string, data []byte) error {
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return errors.New("invalid storage key")
	}
	full := filepath.Join(root, filepath.FromSlash(key))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absFull, absRoot+string(os.PathSeparator)) && absFull != absRoot {
		return errors.New("storage key escapes storage root")
	}
	if err := os.MkdirAll(filepath.Dir(absFull), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absFull, data, 0o644)
}

// openDocumentFile resolves a relative storage key to an absolute path and
// verifies it stays inside the storage root. It returns the path and the
// stored filename (basename) for Content-Disposition.
func openDocumentFile(root, key string) (string, string, error) {
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return "", "", errors.New("invalid storage key")
	}
	full := filepath.Join(root, filepath.FromSlash(key))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", err
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return "", "", err
	}
	if !strings.HasPrefix(absFull, absRoot+string(os.PathSeparator)) && absFull != absRoot {
		return "", "", errors.New("storage key escapes storage root")
	}
	if _, err := os.Stat(absFull); err != nil {
		return "", "", err
	}
	return absFull, filepath.Base(absFull), nil
}

// deleteDocumentFile removes the physical file for a relative storage key.
// It verifies the key stays inside the storage root and tolerates a missing
// file (nothing to clean). Removing the file is safe on soft delete because a
// soft-deleted document row can never be downloaded again.
func deleteDocumentFile(root, key string) error {
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return errors.New("invalid storage key")
	}
	full := filepath.Join(root, filepath.FromSlash(key))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absFull, absRoot+string(os.PathSeparator)) && absFull != absRoot {
		return errors.New("storage key escapes storage root")
	}
	if err := os.Remove(absFull); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
