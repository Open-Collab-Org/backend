## Session keys

Session tokens are stored like the following:

Key | Value
----|------
`session:<session_token>:user.id` | `<user_id>`
`user:<user_id>:session.keys` | `[<session_token>]`

The first key (`session:<session_token>:user.id`) is used to check whether a
session token exists or not and to get the user to which the session token
belongs. The second key `user:<user_id>:session.keys` is a reverse index
that stores all keys a user owns. This is needed when we want to invalidate
all of a user's sessions.

Example:

For a user of id `12` with two valid sessions, there are the following
key-value pairs in redis:
```
GET session:2a5de6a1-5318-47be-a2c8-669ba4402b8c:user.id
12

GET session:027b032f-0d64-4611-9039-ef03bc62ba6e:user.id
12

SMEMBERS user:12:session.keys
1) "2a5de6a1-5318-47be-a2c8-669ba4402b8c"
2) "027b032f-0d64-4611-9039-ef03bc62ba6e"
```




