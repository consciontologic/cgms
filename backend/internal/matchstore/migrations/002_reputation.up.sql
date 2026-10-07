CREATE TABLE reputation_outcomes (
 match_id text NOT NULL REFERENCES matches(match_id) ON DELETE CASCADE,
 game_id text NOT NULL,
 payer integer NOT NULL,
 recipient integer NOT NULL,
 outcome text NOT NULL CHECK(outcome IN ('trust','anyhoo','scam')),
 PRIMARY KEY(match_id,game_id,payer,recipient),
 CHECK(payer<>recipient)
);
