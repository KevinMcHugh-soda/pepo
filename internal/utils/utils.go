package utils

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

// StringUtils contains string utility functions
type StringUtils struct{}

// IsEmpty checks if a string is empty or contains only whitespace
func (StringUtils) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Truncate truncates a string to a maximum length, adding ellipsis if needed
func (StringUtils) Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// TitleCase converts a string to title case
func (StringUtils) TitleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// HTTPUtils contains HTTP utility functions
type HTTPUtils struct{}

// ExtractIDFromPath extracts an ID from a URL path like "/forms/persons/delete/123"
func (HTTPUtils) ExtractIDFromPath(urlPath, prefix string) (string, error) {
	if !strings.HasPrefix(urlPath, prefix) {
		return "", fmt.Errorf("path does not start with prefix %s", prefix)
	}

	id := strings.TrimPrefix(urlPath, prefix)
	id = strings.Trim(id, "/")

	if id == "" {
		return "", fmt.Errorf("no ID found in path")
	}

	return id, nil
}

// WriteJSONError writes a JSON error response
func (HTTPUtils) WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := fmt.Sprintf(`{"error":"%s","status":%d}`, message, statusCode)
	w.Write([]byte(response))
}

// WriteJSONResponse writes a JSON response
func (HTTPUtils) WriteJSONResponse(w http.ResponseWriter, statusCode int, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(data))
}

// GetFormValue safely gets a form value with trimming
func (HTTPUtils) GetFormValue(r *http.Request, key string) string {
	return strings.TrimSpace(r.FormValue(key))
}

// TimeUtils contains time utility functions
type TimeUtils struct{}

// FormatRelative formats a time relative to now (e.g., "2 hours ago")
func (TimeUtils) FormatRelative(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("Jan 2, 2006")
	}
}

// FormatDateTime formats a time for display
func (TimeUtils) FormatDateTime(t time.Time) string {
	return t.Format("Jan 2, 2006 at 3:04 PM")
}

// FormatDate formats a date for display
func (TimeUtils) FormatDate(t time.Time) string {
	return t.Format("Jan 2, 2006")
}

// ValidationUtils contains validation utility functions
type ValidationUtils struct{}

// ValidateRequired checks if required fields are present
func (ValidationUtils) ValidateRequired(fields map[string]string) []string {
	var errors []string
	for field, value := range fields {
		if strings.TrimSpace(value) == "" {
			errors = append(errors, fmt.Sprintf("%s is required", field))
		}
	}
	return errors
}

// ValidateStringLength validates string length constraints
func (ValidationUtils) ValidateStringLength(value, fieldName string, min, max int) error {
	length := len(strings.TrimSpace(value))
	if length < min {
		return fmt.Errorf("%s must be at least %d characters", fieldName, min)
	}
	if max > 0 && length > max {
		return fmt.Errorf("%s must be no more than %d characters", fieldName, max)
	}
	return nil
}

// ConversionUtils contains type conversion utilities
type ConversionUtils struct{}

// StringToInt safely converts string to int with default value
func (ConversionUtils) StringToInt(s string, defaultVal int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultVal
}

// BoolToString converts bool to string
func (ConversionUtils) BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Global utility instances for easy access
var (
	String     = StringUtils{}
	HTTP       = HTTPUtils{}
	Time       = TimeUtils{}
	Validation = ValidationUtils{}
	Convert    = ConversionUtils{}
)

// SafeFileName sanitizes a filename by removing dangerous characters
func SafeFileName(name string) string {
	// Remove or replace dangerous characters
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "..", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")

	// Trim whitespace and limit length
	name = strings.TrimSpace(name)
	if len(name) > 255 {
		ext := path.Ext(name)
		base := name[:255-len(ext)]
		name = base + ext
	}

	return name
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Map applies a function to each element of a slice
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter filters a slice based on a predicate function
func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
