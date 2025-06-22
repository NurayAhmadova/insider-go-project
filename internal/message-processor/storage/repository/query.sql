-- theoretically should be paginated
-- name: ListSentMessages :many
SELECT * FROM messages WHERE sent = true;

-- name: ListUnsentMessages :many
SELECT * FROM messages WHERE sent = false ORDER BY created_at DESC FOR UPDATE SKIP LOCKED LIMIT $1;

-- name: UpdateSentStatus :exec
UPDATE messages m SET sent=true WHERE m.id = @id;