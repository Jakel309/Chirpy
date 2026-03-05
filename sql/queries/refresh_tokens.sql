-- name: CreateRefreshToken :one
insert into refreshTokens(token, created_at, updated_at, user_id, expires_at, revoked_at)
values (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    null
)
RETURNING *;

-- name: GetRefreshToken :one
select * from refreshTokens
where token = $1;

-- name: RevokeToken :one
update refreshTokens
set revoked_at = NOW(), 
updated_at = NOW()
where token = $1
RETURNING *;

-- name: DeleteRefreshTokens :one
delete from refreshTokens
RETURNING *;