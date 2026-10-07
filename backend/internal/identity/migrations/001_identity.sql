CREATE TABLE identity_accounts (
 account_id text PRIMARY KEY,
 kind text NOT NULL CHECK(kind IN ('guest','member','deleted')),
 username text UNIQUE,
 password_hash bytea,
 CHECK ((kind='member' AND username IS NOT NULL AND password_hash IS NOT NULL) OR (kind IN ('guest','deleted') AND username IS NULL AND password_hash IS NULL))
);
CREATE TABLE identity_sessions (
 token_hash bytea PRIMARY KEY CHECK(octet_length(token_hash)=32),
 account_id text NOT NULL REFERENCES identity_accounts,
 expires_at timestamptz NOT NULL
);
CREATE INDEX identity_sessions_account ON identity_sessions(account_id);
