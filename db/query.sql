-- name: ListTodos :many
SELECT * FROM todo;

-- name: AddTodo :one
INSERT INTO todo (title) VALUES (?) RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todo WHERE id = ?;

-- name: UpdateTodo :one
UPDATE todo SET completed = ? WHERE id = ? RETURNING *;
