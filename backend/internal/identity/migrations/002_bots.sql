-- Server-owned opponents have no login credentials. Existing identities retain
-- their original constraints and all previously deployed migration bytes.
ALTER TABLE identity_accounts DROP CONSTRAINT identity_accounts_kind_check;
ALTER TABLE identity_accounts DROP CONSTRAINT identity_accounts_check;
ALTER TABLE identity_accounts ADD CONSTRAINT identity_accounts_kind_check
 CHECK(kind IN ('guest','member','deleted','bot'));
ALTER TABLE identity_accounts ADD CONSTRAINT identity_accounts_check
 CHECK ((kind='member' AND username IS NOT NULL AND password_hash IS NOT NULL)
 OR (kind IN ('guest','deleted','bot') AND username IS NULL AND password_hash IS NULL));
