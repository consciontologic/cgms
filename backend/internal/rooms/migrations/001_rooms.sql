CREATE TABLE rooms (
 room_id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES identity_accounts(account_id),
 capacity smallint NOT NULL CHECK(capacity IN (3,4)),
 games integer NOT NULL CHECK(games BETWEEN 1 AND 100),
 status text NOT NULL CHECK(status IN ('open','started')),
 match_id text UNIQUE REFERENCES matches(match_id),
 CHECK((status='open' AND match_id IS NULL) OR (status='started' AND match_id IS NOT NULL))
);
CREATE TABLE room_members (
 room_id text NOT NULL REFERENCES rooms,
 actor_id text NOT NULL REFERENCES identity_accounts(account_id),
 seat smallint NOT NULL CHECK(seat BETWEEN 1 AND 4),
 PRIMARY KEY(room_id,actor_id), UNIQUE(room_id,seat)
);
CREATE TABLE room_invitations (
 room_id text PRIMARY KEY REFERENCES rooms,
 token_hash bytea NOT NULL UNIQUE CHECK(octet_length(token_hash)=32),
 expires_at timestamptz NOT NULL
);
CREATE TABLE room_commands (
 actor_id text NOT NULL REFERENCES identity_accounts(account_id),
 command_id text NOT NULL,
 body_hash bytea NOT NULL,
 result bytea NOT NULL,
 PRIMARY KEY(actor_id,command_id)
);
