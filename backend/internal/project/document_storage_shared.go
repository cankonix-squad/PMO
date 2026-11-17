package project

// Exported wrappers so other modules (e.g. the contractor portal) store files
// with the same allowlist, MIME sniffing, and path-escape guards as project
// documents, under the same storage root.

// SanitizeUploadFilename returns a safe basename for an uploaded file.
func SanitizeUploadFilename(name string) string { return sanitizeFilename(name) }

// ValidateUploadFile checks the extension allowlist and sniffs the MIME type.
func ValidateUploadFile(header []byte, filename string) (string, error) {
	return validateDocumentFile(header, filename)
}

// SaveStoredFile writes data under root at the relative key.
func SaveStoredFile(root, key string, data []byte) error { return saveDocumentFile(root, key, data) }

// OpenStoredFile resolves a relative key to an absolute path inside root.
func OpenStoredFile(root, key string) (string, string, error) { return openDocumentFile(root, key) }

// DeleteStoredFile removes the file at the relative key.
func DeleteStoredFile(root, key string) error { return deleteDocumentFile(root, key) }
