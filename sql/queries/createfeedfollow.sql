-- name: CreateFeedFollow :one
WITH
  new_follow AS (
    INSERT INTO
      feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES
      ($1, $2, $3, $4, $5)
    RETURNING
      *
  )
SELECT
  new_follow.*,
  users.name AS follower,
  feeds.name AS feed
FROM
  new_follow
  INNER JOIN users ON new_follow.user_id = users.id
  INNER JOIN feeds ON new_follow.feed_id = feeds.id;
