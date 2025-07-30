# Pepo Actions Feature Demo

This document demonstrates the newly implemented Actions feature in the Pepo performance tracking application.

## Overview

The Actions feature allows you to record specific actions that people have performed, with the following attributes:
- **Person**: Who performed the action (linked to existing persons)
- **Description**: What they did
- **Valence**: Whether it was positive or negative
- **Occurred At**: When it happened (defaults to now)
- **References**: Optional links or references related to the action

## Database Schema

### Action Table
```sql
CREATE TABLE action (
    id BYTEA PRIMARY KEY,
    person_id BYTEA NOT NULL REFERENCES person(id) ON DELETE CASCADE,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    description TEXT NOT NULL,
    "references" TEXT,
    valence valence_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT action_description_check CHECK (LENGTH(TRIM(BOTH FROM description)) > 0)
);

CREATE TYPE valence_type AS ENUM ('positive', 'negative');
```

## API Endpoints

### Actions CRUD
- `GET /api/v1/actions` - List all actions (supports filtering)
- `POST /api/v1/actions` - Create a new action
- `GET /api/v1/actions/{id}` - Get specific action
- `PUT /api/v1/actions/{id}` - Update action
- `DELETE /api/v1/actions/{id}` - Delete action

### Person-specific Actions
- `GET /api/v1/persons/{id}/actions` - Get all actions for a person

### Query Parameters
- `limit` - Number of results (default: 20, max: 100)
- `offset` - Pagination offset (default: 0)
- `person_id` - Filter by person ID
- `valence` - Filter by positive/negative

## Example API Usage

### 1. Create a Person
```bash
curl -X POST http://localhost:8080/api/v1/persons \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe"}'
```

Response:
```json
{
  "id": "d258ifuv0le41h0toplg",
  "name": "John Doe",
  "created_at": "2025-07-30T21:00:00Z",
  "updated_at": "2025-07-30T21:00:00Z"
}
```

### 2. Record a Positive Action
```bash
curl -X POST http://localhost:8080/api/v1/actions \
  -H "Content-Type: application/json" \
  -d '{
    "person_id": "d258ifuv0le41h0toplg",
    "description": "Completed project ahead of schedule",
    "valence": "positive",
    "references": "https://github.com/company/project/pull/123"
  }'
```

Response:
```json
{
  "id": "d258jg2v0le41h0topmi",
  "person_id": "d258ifuv0le41h0toplg",
  "occurred_at": "2025-07-30T21:05:00Z",
  "description": "Completed project ahead of schedule",
  "references": "https://github.com/company/project/pull/123",
  "valence": "positive",
  "created_at": "2025-07-30T21:05:00Z",
  "updated_at": "2025-07-30T21:05:00Z"
}
```

### 3. Record a Negative Action
```bash
curl -X POST http://localhost:8080/api/v1/actions \
  -H "Content-Type: application/json" \
  -d '{
    "person_id": "d258ifuv0le41h0toplg",
    "description": "Missed important deadline",
    "valence": "negative",
    "occurred_at": "2025-07-29T14:30:00Z"
  }'
```

### 4. List All Actions
```bash
curl http://localhost:8080/api/v1/actions
```

### 5. Filter Actions by Person
```bash
curl "http://localhost:8080/api/v1/persons/d258ifuv0le41h0toplg/actions"
```

### 6. Filter Actions by Valence
```bash
curl "http://localhost:8080/api/v1/actions?valence=positive"
curl "http://localhost:8080/api/v1/actions?valence=negative"
```

## Web Interface

The web interface at `http://localhost:8080` now includes:

### Enhanced Home Page
- **Two-column layout** with person and action creation forms side by side
- **Person form** on the left (unchanged)
- **Action form** on the right with:
  - Person dropdown (dynamically loaded)
  - Description textarea
  - Valence selector (positive/negative)
  - Optional date/time picker
  - Optional references URL field

### Action Display
- Actions list shows recent actions with:
  - Color-coded valence indicators (green dot for positive, red for negative)
  - Person ID reference
  - Description
  - Clickable reference links
  - Occurred and created timestamps
  - Delete functionality

### HTMX Integration
- **Real-time updates**: New actions appear immediately without page refresh
- **Dynamic person loading**: Person dropdown populated via HTMX
- **Inline deletion**: Actions can be deleted with confirmation
- **Form validation**: Client and server-side validation

## Sample Workflow

1. **Visit** `http://localhost:8080`
2. **Add a person**: Enter name in left form, click "Add Person"
3. **Record positive action**:
   - Select person from dropdown
   - Enter: "Led successful team presentation"
   - Choose "Positive"
   - Add reference: "https://docs.company.com/presentation"
   - Click "Record Action"
4. **Record negative action**:
   - Select same person
   - Enter: "Late to three meetings this week"
   - Choose "Negative"
   - Click "Record Action"
5. **View results**: Both actions appear in the right panel with appropriate color coding

## Technical Features

### Database Optimizations
- **XID storage**: Actions use optimized bytea storage for IDs (40% space savings)
- **Indexes**: Optimized for common queries (person_id, occurred_at, valence)
- **Cascading deletes**: Actions automatically deleted when person is removed
- **Data validation**: Constraints ensure data integrity

### API Features
- **Type-safe**: Generated from OpenAPI specification
- **Filtering**: Multiple filter combinations supported
- **Pagination**: Efficient handling of large datasets
- **Error handling**: Proper HTTP status codes and error messages

### Security & Performance
- **SQL injection protection**: Using parameterized queries via sqlc
- **Connection pooling**: Configured database connection limits
- **Graceful shutdown**: Proper server lifecycle management
- **Input validation**: Both client and server-side validation

## Future Enhancements

Potential improvements for the actions feature:
1. **Action categories**: Beyond just positive/negative
2. **Action templates**: Pre-defined common actions
3. **Bulk actions**: Record multiple actions at once
4. **Action analytics**: Charts and trends over time
5. **Action notifications**: Alerts for significant actions
6. **Action search**: Full-text search across descriptions
7. **Action history**: Track edits and changes
8. **Action approval**: Workflow for sensitive actions

## Testing

Run the test suite:
```bash
make test-api      # Test API endpoints
make test-forms    # Test web forms
./test_actions_api.sh  # Comprehensive action testing
```

The implementation is production-ready with comprehensive error handling, validation, and testing coverage.