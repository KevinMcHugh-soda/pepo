-- name: CreateConversation :one
INSERT INTO conversation (id, person_id, description, occurred_at)
VALUES (
    x2b(sqlc.arg(id)),
    x2b(sqlc.arg(person_id)),
    sqlc.arg(description),
    sqlc.arg(occurred_at)
)
RETURNING sqlc.embed(conversation);
