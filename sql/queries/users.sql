-- name: CreateUser :one
insert into users (id, created_at, updated_at, email, hashed_password)
values (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteUsers :one
delete from users
RETURNING *;

-- name: GetUser :one
select * from users
where email = $1;

-- name: GetUserById :one
select * from users
where id = $1;

-- name: UpdateUser :one
update users
set email = $1,
hashed_password = $2,
updated_at = NOW()
where id = $3
RETURNING *;