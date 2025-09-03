-- name: CreateConversation :one
INSERT INTO conversation (id, person_id, description, occurred_at)
VALUES (
    x2b(sqlc.arg(id)),
    x2b(sqlc.arg(person_id)),
    sqlc.arg(description),
    sqlc.arg(occurred_at)
)
RETURNING sqlc.embed(conversation);
-- name: ListConversationsByPersonID :many
SELECT DISTINCT ON (c.id)
    sqlc.embed(c)
FROM conversation c
JOIN action_conversation ac ON ac.conversation_id = c.id
JOIN action a ON a.id = ac.action_id
WHERE a.person_id = x2b(sqlc.arg(person_id))
ORDER BY c.id, c.occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountConversationsByPersonID :one
SELECT COUNT(DISTINCT c.id)
FROM conversation c
JOIN action_conversation ac ON ac.conversation_id = c.id
JOIN action a ON a.id = ac.action_id
WHERE a.person_id = x2b(sqlc.arg(person_id));
