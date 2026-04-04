package handler

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// parseUUID converts a string to pgtype.UUID
func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(s)
	return id, err
}

// optionalString returns a pgtype.Text for optional string values
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// optionalStringPgtype converts an optional string to pgtype.Text
func optionalStringPgtype(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// optionalStringPgtype2 converts a pointer to string to pgtype.Text (nil pointer = NULL)
func optionalStringPgtype2(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// derefString dereferences a pointer to string, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
