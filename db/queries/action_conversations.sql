-- name: AddActionToConversation :exec
INSERT INTO action_conversation (action_id, conversation_id)
VALUES (x2b(sqlc.arg(action_id)), x2b(sqlc.arg(conversation_id)));
