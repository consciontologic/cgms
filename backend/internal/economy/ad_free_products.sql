-- Fixed-term standalone ad removal is independent of unlimited online starts.
ALTER TABLE economy_provider_purchases DROP CONSTRAINT economy_provider_purchases_kind_check;
ALTER TABLE economy_provider_purchases ADD CONSTRAINT economy_provider_purchases_kind_check CHECK(kind IN ('premium','cosmetic','weekly','monthly','yearly','ad_free'));
