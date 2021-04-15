## Session keys

Session keys are stored like the following:

Key | Value
----|------
`session:<session_key>:user.id` | `<user_id>`
`user:<user_id>:session.keys` | `[<session_key>]`

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


