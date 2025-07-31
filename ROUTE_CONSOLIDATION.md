# Route Consolidation Documentation

## Overview

This document describes the route consolidation work that eliminated duplicate routes by implementing content negotiation. The API now serves both JSON and HTML responses from the same endpoints based on the client's `Accept` header.

## What Changed

### Before: Duplicate Routes

Previously, the application had separate routes for JSON and HTML responses:

```
# JSON API routes
GET  /api/v1/people          → JSON response
POST /api/v1/people          → JSON response
GET  /api/v1/people/{id}     → JSON response
PUT  /api/v1/people/{id}     → JSON response
DELETE /api/v1/people/{id}   → 204 No Content

# HTML Form routes (for HTMX)
GET  /forms/people/list      → HTML response
POST /forms/people/create    → HTML response
DELETE /forms/people/delete/{id} → HTML response
GET  /forms/people/select    → HTML response
```

### After: Consolidated Routes with Content Negotiation

Now, the same endpoints serve both content types:

```
# Consolidated routes supporting both JSON and HTML
GET  /api/v1/people          → JSON or HTML (based on Accept header)
POST /api/v1/people          → JSON or HTML (based on Accept header)
GET  /api/v1/people/{id}     → JSON or HTML (based on Accept header)
PUT  /api/v1/people/{id}     → JSON or HTML (based on Accept header)
DELETE /api/v1/people/{id}   → 204 No Content

# Convenience routes (optional, same functionality)
GET  /people          → JSON or HTML (based on Accept header)
POST /people          → JSON or HTML (based on Accept header)
# ... etc
```

## How Content Negotiation Works

### Accept Header Handling

The server examines the `Accept` header in the HTTP request:

- `Accept: application/json` → Returns JSON response
- `Accept: text/html` → Returns HTML response
- No Accept header → Returns JSON response (default)

### Example Requests

#### JSON Request
```bash
curl -H "Accept: application/json" http://localhost:8000/api/v1/people
```
Response:
```json
{
  "persons": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "John Doe",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

#### HTML Request
```bash
curl -H "Accept: text/html" http://localhost:8000/api/v1/people
```
Response:
```html
<div class="person-list">
  <div class="person-item" data-id="123e4567-e89b-12d3-a456-426614174000">
    <span class="person-name">John Doe</span>
    <time>2023-01-01T00:00:00Z</time>
  </div>
</div>
```

## Implementation Details

### Architecture Components

1. **ContentNegotiatingHandler** (`internal/handlers/content_negotiation.go`)
   - Wraps existing business logic handlers
   - Examines HTTP request context for Accept header
   - Returns appropriate response type (JSON vs HTML)

2. **Request Context Middleware** (`internal/middleware/context.go`)
   - Adds HTTP request to context for access by handlers
   - Enables content negotiation without changing handler signatures

3. **Updated OpenAPI Specification** (`api/openapi.yaml`)
   - All successful responses now support both `application/json` and `text/html`
   - Single source of truth for API endpoints

4. **Template Integration**
   - Reuses existing templ templates for HTML responses
   - Converts API response structures to template structures

### Response Type Mapping

| HTTP Status | JSON Response Type | HTML Response Type |
|-------------|-------------------|-------------------|
| 200 (GET)   | `GetPersonsOKApplicationJSON` | `GetPersonsOKTextHTML` |
| 201 (POST)  | `Person` | `CreatePersonCreatedTextHTML` |
| 200 (PUT)   | `Person` | `UpdatePersonOKTextHTML` |
| 204 (DELETE)| No content | No content |

## Benefits

### 1. **Reduced Code Duplication**
- Eliminated separate form handlers
- Single business logic path for all requests
- Consistent validation and error handling

### 2. **API-First Design**
- All functionality available via REST API
- HTML rendering is just a presentation layer
- Easy to add new content types (XML, CSV, etc.)

### 3. **Better Testing**
- Single set of endpoints to test
- Content negotiation can be tested independently
- Reduced test maintenance

### 4. **Improved Developer Experience**
- Consistent URL structure
- Same endpoints for web browsers and API clients
- Clear separation of concerns

### 5. **Future-Proof**
- Easy to add new content types
- Ready for GraphQL or other API patterns
- Scalable architecture

## Migration Guide

### For API Clients

No changes required for existing JSON API clients. The `/api/v1/*` endpoints continue to work exactly as before.

### For HTML/HTMX Clients

Update requests to use the consolidated endpoints:

#### Before:
```javascript
// HTMX
<form hx-post="/forms/people/create" hx-target="#person-list">
  <input name="name" type="text" required>
  <button type="submit">Create Person</button>
</form>
```

#### After:
```javascript
// HTMX with Accept header
<form hx-post="/api/v1/people" 
      hx-headers='{"Accept": "text/html"}' 
      hx-target="#person-list">
  <input name="name" type="text" required>
  <button type="submit">Create Person</button>
</form>
```

### For New Clients

Use the convenience routes for cleaner URLs:

```javascript
// Even cleaner
<form hx-post="/people" 
      hx-headers='{"Accept": "text/html"}' 
      hx-target="#person-list">
  <input name="name" type="text" required>
  <button type="submit">Create Person</button>
</form>
```

## Legacy Support

The old `/forms/*` routes are still available but marked as deprecated:

- `/forms/people/create` → Use `POST /people` with `Accept: text/html`
- `/forms/people/list` → Use `GET /people` with `Accept: text/html`
- `/forms/people/delete/{id}` → Use `DELETE /people/{id}`
- `/forms/actions/create` → Use `POST /actions` with `Accept: text/html`
- `/forms/actions/list` → Use `GET /actions` with `Accept: text/html`
- `/forms/actions/delete/{id}` → Use `DELETE /actions/{id}`

**Recommendation**: Migrate to the consolidated routes to take advantage of the improved architecture.

## Testing

Use the test script to verify content negotiation:

```bash
./test/test_consolidated_routes.sh
```

This script tests:
- JSON responses with `Accept: application/json`
- HTML responses with `Accept: text/html`
- Default behavior (no Accept header)
- All CRUD operations for people and actions
- Error responses in both formats

## Configuration

No additional configuration is required. Content negotiation is enabled by default.

### Environment Variables

All existing environment variables work unchanged:
- `PORT` - Server port (default: 8000)
- `DATABASE_URL` - Database connection string
- `ENVIRONMENT` - Environment mode (development/production)

## Performance Considerations

### Minimal Overhead
- Content negotiation adds negligible latency
- Same business logic execution path
- Template rendering only when HTML is requested

### Caching
- JSON responses can be cached normally
- HTML responses should not be cached (dynamic content)
- CDN can cache based on Accept header

## Security

### Content Type Validation
- Only `application/json` and `text/html` are supported
- Invalid Accept headers default to JSON
- No risk of content type confusion attacks

### HTMX Considerations
- HTML responses are safe for HTMX consumption
- No additional XSS vectors introduced
- Same CSRF protection applies

## Future Enhancements

### Planned Improvements
1. **Additional Content Types**
   - CSV export (`Accept: text/csv`)
   - XML responses (`Accept: application/xml`)
   - PDF reports (`Accept: application/pdf`)

2. **Enhanced HTML Responses**
   - Full page responses for browser navigation
   - Partial responses for HTMX updates
   - Progressive enhancement support

3. **Content Negotiation Extensions**
   - Quality factors (`Accept: application/json;q=0.9, text/html;q=0.8`)
   - Language negotiation
   - Encoding negotiation

### Migration Timeline
- **Phase 1** (Complete): Content negotiation implementation
- **Phase 2** (Next): HTMX migration to consolidated routes
- **Phase 3** (Future): Deprecate legacy `/forms/*` routes
- **Phase 4** (Future): Additional content type support

## Troubleshooting

### Common Issues

1. **HTML responses not rendering**
   - Verify `Accept: text/html` header is set
   - Check browser developer tools for request headers

2. **HTMX not working with new routes**
   - Add `hx-headers='{"Accept": "text/html"}'` to HTMX elements
   - Verify content type in response headers

3. **JSON responses when expecting HTML**
   - Check Accept header spelling and case
   - Verify middleware is properly configured

### Debug Information

Enable debug logging to see content negotiation decisions:

```bash
ENVIRONMENT=development ./server
```

Look for log entries showing Accept header processing and response type selection.

## Conclusion

The route consolidation successfully eliminates code duplication while maintaining full backward compatibility. The new architecture is more maintainable, testable, and future-proof while providing the same functionality as the previous separate route system.

For questions or issues, refer to the test scripts or examine the implementation in `internal/handlers/content_negotiation.go`.