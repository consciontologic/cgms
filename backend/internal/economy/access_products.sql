-- Preserve the original ad-free promise once; new earned time never extends it.
ALTER TABLE economy_accounts ADD COLUMN legacy_ad_free_until timestamptz NOT NULL DEFAULT '1970-01-01T00:00:00Z';
UPDATE economy_accounts SET legacy_ad_free_until=pass_until;
-- Retired durations remain valid historical receipts, not new product choices.
ALTER TABLE economy_pass_operations DROP CONSTRAINT economy_pass_operations_days_check;
ALTER TABLE economy_pass_operations ADD CONSTRAINT economy_pass_operations_days_check CHECK(days BETWEEN 1 AND 7);
ALTER TABLE economy_provider_purchases DROP CONSTRAINT economy_provider_purchases_kind_check;
ALTER TABLE economy_provider_purchases ADD CONSTRAINT economy_provider_purchases_kind_check CHECK(kind IN ('premium','cosmetic','weekly','monthly','yearly'));
