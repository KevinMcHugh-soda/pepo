-- name: CreateConversation :one
INSERT INTO conversation (id, description, occurred_at)
VALUES (
    x2b(sqlc.arg(id)),
    sqlc.arg(description),
    sqlc.arg(occurred_at)
)
RETURNING sqlc.embed(conversation);
