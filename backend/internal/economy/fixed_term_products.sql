-- Approved Daily grants share paid unlimited access and ad removal.
-- Confirmation instants remain in immutable transaction grant_state metadata.
ALTER TABLE economy_provider_purchases DROP CONSTRAINT economy_provider_purchases_kind_check;
ALTER TABLE economy_provider_purchases ADD CONSTRAINT economy_provider_purchases_kind_check CHECK(kind IN ('premium','cosmetic','daily','weekly','monthly','yearly','ad_free'));
