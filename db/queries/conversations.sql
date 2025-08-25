-- name: CreateConversation :one
INSERT INTO conversation (id, description, occurred_at)
VALUES (x2b(sqlc.arg(id)), sqlc.arg(description), sqlc.arg(occurred_at))
RETURNING conversation.id, conversation.description, conversation.occurred_at, conversation.created_at, conversation.updated_at;

-- name: AddActionToConversation :exec
INSERT INTO action_conversation (action_id, conversation_id)
VALUES (x2b(sqlc.arg(action_id)), x2b(sqlc.arg(conversation_id)));

-- name: AddThemeToConversation :exec
INSERT INTO conversation_theme (conversation_id, theme_id)
VALUES (x2b(sqlc.arg(conversation_id)), x2b(sqlc.arg(theme_id)));
