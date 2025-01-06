-- name: ListTodos :many
SELECT * FROM todo;

-- name: AddTodo :one
INSERT INTO todo (title) VALUES (?) RETURNING *;
