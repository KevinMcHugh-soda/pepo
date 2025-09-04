-- name: AddActionToConversation :exec
INSERT INTO action_conversation (action_id, conversation_id)
VALUES (x2b(sqlc.arg(action_id)), x2b(sqlc.arg(conversation_id)));

-- name: ListActionIDsByConversationID :many
SELECT action_id
FROM action_conversation
WHERE conversation_id = x2b(sqlc.arg(conversation_id));

-- name: DeleteActionsByConversationID :exec
DELETE FROM action_conversation
WHERE conversation_id = x2b(sqlc.arg(conversation_id));
