package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// FormToJSONAdapter middleware converts HTML form submissions to JSON requests
type FormToJSONAdapter struct {
	next http.Handler
}

// NewFormToJSONAdapter creates a new form-to-JSON adapter middleware
func NewFormToJSONAdapter(next http.Handler) *FormToJSONAdapter {
	return &FormToJSONAdapter{next: next}
}

func (f *FormToJSONAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only process POST and PUT requests with form data
	if (r.Method == "POST" || r.Method == "PUT") && f.isFormData(r) {
		// Convert form data to JSON
		jsonData, err := f.convertFormToJSON(r)
		if err != nil {
			http.Error(w, "Invalid form data: "+err.Error(), http.StatusBadRequest)
			return
		}
		if jsonData != nil {
			// Create a new request with JSON data
			newRequest := f.createJSONRequest(r, jsonData)

			// Pass the modified request to the next handler
			f.next.ServeHTTP(w, newRequest)
			return
		}

		// If no conversion happened, fall through and pass original request
	}

	// Pass through non-form requests unchanged
	f.next.ServeHTTP(w, r)
}

// isFormData checks if the request contains form data
func (f *FormToJSONAdapter) isFormData(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.HasPrefix(contentType, "multipart/form-data")
}

// convertFormToJSON converts form data to JSON based on the URL path
func (f *FormToJSONAdapter) convertFormToJSON(r *http.Request) ([]byte, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	// Determine the type of form based on URL path
	switch {
	case strings.HasPrefix(r.URL.Path, "/people"):
		return f.convertPersonForm(r)
	case strings.HasPrefix(r.URL.Path, "/actions"):
		return f.convertActionForm(r)
	case strings.Contains(r.URL.Path, "/conversations"):
		return f.convertConversationForm(r)
	case strings.HasPrefix(r.URL.Path, "/forms/themes"):
		return f.convertThemeForm(r)
	default:
		// Unknown form type - skip conversion
		return nil, nil
	}
}

// convertPersonForm converts person form data to JSON
func (f *FormToJSONAdapter) convertPersonForm(r *http.Request) ([]byte, error) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return nil, &FormError{Field: "name", Message: "Name is required"}
	}

	data := map[string]interface{}{
		"name": name,
	}

	return json.Marshal(data)
}

// convertActionForm converts action form data to JSON
func (f *FormToJSONAdapter) convertActionForm(r *http.Request) ([]byte, error) {
	// Required fields
	personID := strings.TrimSpace(r.FormValue("person_id"))
	if personID == "" {
		return nil, &FormError{Field: "person_id", Message: "Person ID is required"}
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		return nil, &FormError{Field: "description", Message: "Description is required"}
	}

	valence := strings.TrimSpace(r.FormValue("valence"))
	if valence == "" {
		return nil, &FormError{Field: "valence", Message: "Valence is required"}
	}

	// Validate valence
	if valence != "positive" && valence != "negative" && valence != "neutral" {
		return nil, &FormError{Field: "valence", Message: "Valence must be positive, negative, or neutral"}
	}

	// Parse occurred_at (optional, defaults to current time if not provided)
	var occurredAt time.Time
	occurredAtStr := strings.TrimSpace(r.FormValue("occurred_at"))
	if occurredAtStr != "" {
		// Try parsing HTML datetime-local format: 2006-01-02T15:04
		var err error
		occurredAt, err = time.Parse("2006-01-02T15:04", occurredAtStr)
		if err != nil {
			// Try with seconds: 2006-01-02T15:04:05
			occurredAt, err = time.Parse("2006-01-02T15:04:05", occurredAtStr)
			if err != nil {
				// Try ISO format
				occurredAt, err = time.Parse(time.RFC3339, occurredAtStr)
				if err != nil {
					return nil, &FormError{Field: "occurred_at", Message: "Invalid date format"}
				}
			}
		}
	} else {
		occurredAt = time.Now()
	}

	// Build JSON data
	data := map[string]interface{}{
		"person_id":   personID,
		"occurred_at": occurredAt.Format(time.RFC3339),
		"description": description,
		"valence":     valence,
	}

	// Optional references field
	references := strings.TrimSpace(r.FormValue("references"))
	if references != "" {
		data["references"] = references
	}

	// Optional themes field
	if themes := r.Form["themes"]; len(themes) > 0 {
		clean := make([]string, 0, len(themes))
		for _, t := range themes {
			t = strings.TrimSpace(t)
			if t != "" {
				clean = append(clean, t)
			}
		}
		if len(clean) > 0 {
			data["themes"] = clean
		}
	}

	return json.Marshal(data)
}

// convertThemeForm converts theme creation form data to JSON
func (f *FormToJSONAdapter) convertThemeForm(r *http.Request) ([]byte, error) {
	personID := strings.TrimSpace(r.FormValue("person_id"))
	if personID == "" {
		return nil, &FormError{Field: "person_id", Message: "Person ID is required"}
	}

	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		return nil, &FormError{Field: "text", Message: "Text is required"}
	}

	data := map[string]interface{}{
		"person_id": personID,
		"text":      text,
	}

	if themes := r.Form["themes"]; len(themes) > 0 {
		clean := make([]string, 0, len(themes))
		for _, t := range themes {
			t = strings.TrimSpace(t)
			if t != "" {
				clean = append(clean, t)
			}
		}
		if len(clean) > 0 {
			data["themes"] = clean
		}
	}

	return json.Marshal(data)
}

// convertConversationForm converts conversation form data to JSON
func (f *FormToJSONAdapter) convertConversationForm(r *http.Request) ([]byte, error) {
	// Required fields
	personID := strings.TrimSpace(r.FormValue("person_id"))
	if personID == "" {
		return nil, &FormError{Field: "person_id", Message: "Person ID is required"}
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		return nil, &FormError{Field: "description", Message: "Description is required"}
	}

	// Parse occurred_at (optional, defaults to current time if not provided)
	var occurredAt time.Time
	occurredAtStr := strings.TrimSpace(r.FormValue("occurred_at"))
	if occurredAtStr != "" {
		// Try parsing HTML datetime-local format: 2006-01-02T15:04
		var err error
		occurredAt, err = time.Parse("2006-01-02T15:04", occurredAtStr)
		if err != nil {
			// Try with seconds: 2006-01-02T15:04:05
			occurredAt, err = time.Parse("2006-01-02T15:04:05", occurredAtStr)
			if err != nil {
				// Try ISO format
				occurredAt, err = time.Parse(time.RFC3339, occurredAtStr)
				if err != nil {
					return nil, &FormError{Field: "occurred_at", Message: "Invalid date format"}
				}
			}
		}
	} else {
		occurredAt = time.Now()
	}

	// Build JSON data
	data := map[string]interface{}{
		"person_id":   personID,
		"occurred_at": occurredAt.Format(time.RFC3339),
		"description": description,
	}

	return json.Marshal(data)
}

// createJSONRequest creates a new request with JSON data
func (f *FormToJSONAdapter) createJSONRequest(r *http.Request, jsonData []byte) *http.Request {
	// Create new request with JSON body
	newRequest := r.Clone(r.Context())
	newRequest.Body = io.NopCloser(bytes.NewReader(jsonData))
	newRequest.ContentLength = int64(len(jsonData))

	// Update headers
	newRequest.Header.Set("Content-Type", "application/json")

	// Preserve Accept header for content negotiation
	// If no Accept header is set, default to text/html for form submissions
	if newRequest.Header.Get("Accept") == "" {
		newRequest.Header.Set("Accept", "text/html")
	}

	return newRequest
}

// FormError represents a form validation error
type FormError struct {
	Field   string
	Message string
}

func (e *FormError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// FormToJSONMiddleware returns a middleware function
func FormToJSONMiddleware(next http.Handler) http.Handler {
	return NewFormToJSONAdapter(next)
}
