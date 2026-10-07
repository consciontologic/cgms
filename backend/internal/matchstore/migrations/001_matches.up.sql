CREATE TABLE matches (
 match_id text PRIMARY KEY CHECK (match_id <> ''),
 snapshot bytea NOT NULL,
 state_version bigint NOT NULL CHECK (state_version >= 0),
 current_game_id text NOT NULL UNIQUE CHECK (current_game_id <> ''),
 config bytea NOT NULL,
 random_state bytea NOT NULL,
 metadata bytea NOT NULL
);
CREATE TABLE members (
 match_id text NOT NULL REFERENCES matches ON DELETE CASCADE,
 actor_id text NOT NULL CHECK (actor_id <> ''),
 seat smallint NOT NULL CHECK (seat BETWEEN 1 AND 4),
 PRIMARY KEY(match_id,actor_id), UNIQUE(match_id,seat)
);
CREATE TABLE command_results (
 match_id text NOT NULL,
 actor_id text NOT NULL,
 command_id text NOT NULL CHECK(command_id <> ''),
 body_hash bytea NOT NULL,
 result bytea NOT NULL,
 PRIMARY KEY(match_id,actor_id,command_id),
 FOREIGN KEY(match_id,actor_id) REFERENCES members ON DELETE CASCADE
);
CREATE TABLE match_events (
 match_id text NOT NULL REFERENCES matches ON DELETE CASCADE,
 seq bigint NOT NULL CHECK(seq > 0),
 game_id text NOT NULL,
 event bytea NOT NULL,
 PRIMARY KEY(match_id,seq)
);
CREATE TABLE seat_events (
 match_id text NOT NULL,
 seat smallint NOT NULL,
 seq bigint NOT NULL,
 projection bytea NOT NULL,
 PRIMARY KEY(match_id,seat,seq),
 FOREIGN KEY(match_id,seat) REFERENCES members(match_id,seat) ON DELETE CASCADE,
 FOREIGN KEY(match_id,seq) REFERENCES match_events ON DELETE CASCADE
);
CREATE TABLE seat_projections (
 match_id text NOT NULL,
 seat smallint NOT NULL,
 state_version bigint NOT NULL CHECK(state_version >= 0),
 cursor bigint NOT NULL CHECK(cursor >= 0),
 projection bytea NOT NULL,
 PRIMARY KEY(match_id,seat),
 FOREIGN KEY(match_id,seat) REFERENCES members(match_id,seat) ON DELETE CASCADE
);
CREATE TABLE games (
 match_id text NOT NULL REFERENCES matches ON DELETE CASCADE,
 game_id text NOT NULL UNIQUE,
 slot integer NOT NULL CHECK(slot >= 0),
 phase text NOT NULL CHECK(phase IN ('playing','settlement','finalized','void')),
 payload bytea NOT NULL,
 PRIMARY KEY(match_id,game_id)
);
-- Canonical string components preserve arbitrary-precision exact rationals.
-- The engine also validates reduction and semantic sign on every read/write.
CREATE DOMAIN exact_numerator AS text CHECK(VALUE ~ '^(0|-?[1-9][0-9]*)$');
CREATE DOMAIN exact_denominator AS text CHECK(VALUE ~ '^[1-9][0-9]*$');
CREATE TABLE game_balances (
 match_id text NOT NULL,
 game_id text NOT NULL,
 seat smallint NOT NULL CHECK(seat BETWEEN 0 AND 3),
 score_num exact_numerator NOT NULL, score_den exact_denominator NOT NULL,
 cash_num exact_numerator NOT NULL CHECK(cash_num !~ '^-'), cash_den exact_denominator NOT NULL,
 PRIMARY KEY(match_id,game_id,seat),
 FOREIGN KEY(match_id,game_id) REFERENCES games ON DELETE CASCADE
);
CREATE TABLE financial_charges (
 match_id text NOT NULL REFERENCES matches ON DELETE CASCADE,
 charge_id text NOT NULL,
 game_id text NOT NULL,
 debtor smallint NOT NULL CHECK(debtor BETWEEN 0 AND 3),
 creditor smallint NOT NULL CHECK(creditor BETWEEN -1 AND 3 AND creditor <> debtor),
 amount_num exact_numerator NOT NULL CHECK(amount_num !~ '^(-|0$)'), amount_den exact_denominator NOT NULL,
 PRIMARY KEY(match_id,charge_id),
 FOREIGN KEY(match_id,game_id) REFERENCES games ON DELETE CASCADE
);
CREATE TABLE financial_debts (
 match_id text NOT NULL,
 debt_id text NOT NULL,
 charge_id text NOT NULL,
 game_id text NOT NULL,
 debtor smallint NOT NULL CHECK(debtor BETWEEN 0 AND 3),
 creditor smallint NOT NULL CHECK(creditor BETWEEN -1 AND 3 AND creditor <> debtor),
 remaining_num exact_numerator NOT NULL CHECK(remaining_num !~ '^-' ), remaining_den exact_denominator NOT NULL,
 sequence bigint NOT NULL CHECK(sequence >= 0),
 PRIMARY KEY(match_id,debt_id), UNIQUE(match_id,sequence),
 FOREIGN KEY(match_id,charge_id) REFERENCES financial_charges ON DELETE CASCADE,
 FOREIGN KEY(match_id,game_id) REFERENCES games ON DELETE CASCADE
);
CREATE TABLE financial_corrections (
 match_id text NOT NULL,
 operation_id text NOT NULL,
 debt_id text NOT NULL,
 charge_id text NOT NULL,
 game_id text NOT NULL,
 debtor smallint NOT NULL CHECK(debtor BETWEEN 0 AND 3),
 amount_num exact_numerator NOT NULL CHECK(amount_num !~ '^(-|0$)'), amount_den exact_denominator NOT NULL,
 match_level boolean NOT NULL,
 PRIMARY KEY(match_id,operation_id),
 FOREIGN KEY(match_id,charge_id) REFERENCES financial_charges ON DELETE CASCADE,
 FOREIGN KEY(match_id,game_id) REFERENCES games ON DELETE CASCADE
);
CREATE TABLE promises (
 match_id text NOT NULL,
 game_id text NOT NULL,
 promise_id text NOT NULL,
 payer smallint NOT NULL CHECK(payer BETWEEN 0 AND 3),
 recipient smallint NOT NULL CHECK(recipient BETWEEN 0 AND 3 AND recipient <> payer),
 award_id text NOT NULL,
 status text NOT NULL,
 payload bytea NOT NULL,
 PRIMARY KEY(match_id,game_id,promise_id),
 FOREIGN KEY(match_id,game_id) REFERENCES games ON DELETE CASCADE
);
CREATE TABLE proposals (
 match_id text NOT NULL,
 game_id text NOT NULL,
 offer_id text NOT NULL,
 revision bigint NOT NULL CHECK(revision > 0),
 origin_turn bigint NOT NULL CHECK(origin_turn >= 0),
 status text NOT NULL,
 payload bytea NOT NULL,
 PRIMARY KEY(match_id,game_id,offer_id),
 FOREIGN KEY(match_id,game_id) REFERENCES games ON DELETE CASCADE
);
CREATE TABLE settlement_workspaces (
 match_id text PRIMARY KEY REFERENCES matches ON DELETE CASCADE,
 settlement_id text NOT NULL,
 version bigint NOT NULL CHECK(version >= 0),
 payload bytea NOT NULL
);

CREATE TABLE server_results (
 match_id text NOT NULL REFERENCES matches ON DELETE CASCADE,
 command_id text NOT NULL,
 body_hash bytea NOT NULL,
 result bytea NOT NULL,
 PRIMARY KEY(match_id,command_id)
);
