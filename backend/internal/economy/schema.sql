CREATE TABLE economy_accounts (
 account_id text PRIMARY KEY REFERENCES identity_accounts,
 dirt bigint NOT NULL DEFAULT 0 CHECK(dirt BETWEEN 0 AND 1000000000000000),
 pass_until timestamptz NOT NULL DEFAULT '1970-01-01T00:00:00Z',
 premium_until timestamptz NOT NULL DEFAULT '1970-01-01T00:00:00Z'
);
CREATE TABLE economy_starts (
 game_id text PRIMARY KEY,
 actors text[] NOT NULL CHECK(cardinality(actors) IN (3,4)),
 started_at timestamptz NOT NULL
);
CREATE TABLE economy_charges (
 game_id text NOT NULL REFERENCES economy_starts,
 account_id text NOT NULL REFERENCES economy_accounts,
 utc_day date NOT NULL,
 charged_free boolean NOT NULL,
 PRIMARY KEY(game_id,account_id)
);
CREATE INDEX economy_daily ON economy_charges(account_id,utc_day) WHERE charged_free;
CREATE TABLE economy_rewards (
 game_id text NOT NULL REFERENCES economy_starts,
 account_id text NOT NULL REFERENCES economy_accounts,
 amount bigint NOT NULL CHECK(amount>0),
 PRIMARY KEY(game_id,account_id)
);
CREATE TABLE economy_pass_operations (
 account_id text NOT NULL REFERENCES economy_accounts,
 operation_id text NOT NULL,
 days integer NOT NULL CHECK(days BETWEEN 1 AND 6),
 cost bigint NOT NULL CHECK(cost>0),
 expires_at timestamptz NOT NULL,
 PRIMARY KEY(account_id,operation_id)
);
CREATE TABLE economy_promotions (
 code_hash bytea PRIMARY KEY CHECK(octet_length(code_hash)=32),
 days integer NOT NULL CHECK(days BETWEEN 1 AND 366),
 total_limit integer NOT NULL CHECK(total_limit BETWEEN 1 AND 1000000),
 per_account_limit integer NOT NULL CHECK(per_account_limit BETWEEN 1 AND total_limit),
 starts_at timestamptz NOT NULL,
 expires_at timestamptz NOT NULL CHECK(expires_at>starts_at),
 members_only boolean NOT NULL,
 redeemed integer NOT NULL DEFAULT 0 CHECK(redeemed BETWEEN 0 AND total_limit)
);
CREATE TABLE economy_promotion_operations (
 account_id text NOT NULL REFERENCES economy_accounts,
 operation_id text NOT NULL,
 code_hash bytea NOT NULL REFERENCES economy_promotions,
 expires_at timestamptz NOT NULL,
 PRIMARY KEY(account_id,operation_id)
);
CREATE INDEX economy_promotion_usage ON economy_promotion_operations(code_hash,account_id);
CREATE TABLE economy_provider_purchases (
 provider text NOT NULL,
 environment text NOT NULL,
 app text NOT NULL,
 transaction_id text NOT NULL,
 account_id text NOT NULL REFERENCES economy_accounts,
 kind text NOT NULL CHECK(kind IN ('premium','cosmetic')),
 cosmetic text NOT NULL,
 status text NOT NULL CHECK(status IN ('purchased','revoked','pending','unknown')),
 expires_at timestamptz NOT NULL,
 grant_state jsonb NOT NULL,
 PRIMARY KEY(provider,environment,app,transaction_id)
);
CREATE INDEX economy_provider_account ON economy_provider_purchases(account_id);
