-- name: CreateAction :one
INSERT INTO action (id, person_id, occurred_at, description, "references", valence)
VALUES (
    x2b(sqlc.arg(id)),
    x2b(sqlc.arg(person_id)),
    sqlc.arg(occurred_at),
    sqlc.arg(description),
    sqlc.arg('references'),
    sqlc.arg(valence)
)
RETURNING sqlc.embed(action);

-- name: GetActionByID :one
SELECT sqlc.embed(action)
FROM action
WHERE id = x2b(sqlc.arg(id));

-- name: ListActions :many
SELECT sqlc.embed(action), person.name as person_name
FROM action
JOIN person ON action.person_id = person.id
ORDER BY occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListActionsByPersonID :many
SELECT sqlc.embed(action)
FROM action
WHERE person_id = x2b(sqlc.arg(person_id))
ORDER BY occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountActions :one
SELECT COUNT(*) FROM action;

-- name: CountActionsByPersonID :one
SELECT COUNT(*) FROM action WHERE person_id = x2b(sqlc.arg(person_id));

-- name: UpdateAction :one
UPDATE action
SET person_id = x2b(sqlc.arg(person_id)),
    occurred_at = sqlc.arg(occurred_at),
    description = sqlc.arg(description),
    "references" = sqlc.arg('references'),
    valence = sqlc.arg(valence),
    updated_at = NOW()
WHERE id = x2b(sqlc.arg(id))
RETURNING sqlc.embed(action);

-- name: DeleteAction :exec
DELETE FROM action
WHERE id = x2b(sqlc.arg(id));

-- name: ListActionsByValence :many
SELECT sqlc.embed(action)
FROM action
WHERE valence = sqlc.arg(valence)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListActionsByPersonIDAndValence :many
SELECT sqlc.embed(action)
FROM action
WHERE person_id = x2b(sqlc.arg(person_id)) AND valence = sqlc.arg(valence)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: SearchActionsByDescription :many
SELECT sqlc.embed(action)
FROM action
WHERE description ILIKE '%' || sqlc.arg('search') || '%'
ORDER BY occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

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
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetActionsByDateRange :many
SELECT sqlc.embed(action)
FROM action
WHERE occurred_at >= sqlc.arg(start_time) AND occurred_at <= sqlc.arg(end_time)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetRecentActionsByPersonID :many
SELECT sqlc.embed(action)
FROM action
WHERE person_id = x2b(sqlc.arg(person_id)) AND occurred_at >= sqlc.arg(since)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
