-- name: CreateAction :one
INSERT INTO action (id, person_id, occurred_at, description, "references", valence)
VALUES (x2b($1), x2b($2), $3, $4, $5, $6)
RETURNING sqlc.embed(action);

-- name: GetActionByID :one
SELECT sqlc.embed(action)
FROM action
WHERE id = x2b($1);

-- name: ListActions :many
SELECT sqlc.embed(action)
FROM action
ORDER BY occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: ListActionsByPersonID :many
SELECT sqlc.embed(action)
FROM action
WHERE person_id = x2b($1)
ORDER BY occurred_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActions :one
SELECT COUNT(*) FROM action;

-- name: CountActionsByPersonID :one
SELECT COUNT(*) FROM action WHERE person_id = x2b($1);

-- name: UpdateAction :one
UPDATE action
SET person_id = x2b($2), occurred_at = $3, description = $4, "references" = $5, valence = $6, updated_at = NOW()
WHERE id = x2b($1)
RETURNING sqlc.embed(action);

-- name: DeleteAction :exec
DELETE FROM action
WHERE id = x2b($1);

-- name: ListActionsByValence :many
SELECT sqlc.embed(action)
FROM action
WHERE valence = $1
ORDER BY occurred_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActionsByPersonIDAndValence :many
SELECT sqlc.embed(action)
FROM action
WHERE person_id = x2b($1) AND valence = $2
ORDER BY occurred_at DESC
LIMIT $3 OFFSET $4;

-- name: SearchActionsByDescription :many
SELECT sqlc.embed(action)
FROM action
WHERE description ILIKE '%' || $1 || '%'
ORDER BY occurred_at DESC
LIMIT $2 OFFSET $3;

-- name: GetActionsWithPersonDetails :many
SELECT
    b2x(a.id) as action_id,
    b2x(a.person_id) as person_id,
    p.name as person_name,
    a.occurred_at,
    a.description,
    a."references",
    a.valence,
    a.created_at,
    a.updated_at
FROM action a
JOIN person p ON a.person_id = p.id
ORDER BY a.occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: GetActionsByDateRange :many
SELECT sqlc.embed(action)
FROM action
WHERE occurred_at >= $1 AND occurred_at <= $2
ORDER BY occurred_at DESC
LIMIT $3 OFFSET $4;

-- name: GetRecentActionsByPersonID :many
SELECT sqlc.embed(action)
FROM action
WHERE person_id = x2b($1) AND occurred_at >= $2
ORDER BY occurred_at DESC
LIMIT $3 OFFSET $4;
