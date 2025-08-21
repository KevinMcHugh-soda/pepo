-- name: CreatePerson :one
INSERT INTO person (id, name)
VALUES (x2b(sqlc.arg(id)), sqlc.arg(name))
RETURNING b2x(id) as id, name, created_at, updated_at;

-- name: GetPersonByID :one
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
WHERE id = x2b(sqlc.arg(id));

-- name: ListPersons :many
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountPersons :one
SELECT COUNT(*) FROM person;

-- name: UpdatePerson :one
UPDATE person
SET name = sqlc.arg(name), updated_at = NOW()
WHERE id = x2b(sqlc.arg(id))
RETURNING b2x(id) as id, name, created_at, updated_at;

-- name: DeletePerson :exec
DELETE FROM person
WHERE id = x2b(sqlc.arg(id));

-- name: GetPersonByName :one
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
WHERE name = sqlc.arg(name);

-- name: SearchPersonsByName :many
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
WHERE name ILIKE '%' || sqlc.arg('search') || '%'
ORDER BY name
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListPersonsWithLastAction :many
SELECT
    b2x(p.id) as id,
    p.name,
    p.created_at,
    p.updated_at,
    MAX(a.occurred_at) as last_action_at
FROM person p
LEFT JOIN action a ON p.id = a.person_id
GROUP BY p.id, p.name, p.created_at, p.updated_at
ORDER BY p.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
