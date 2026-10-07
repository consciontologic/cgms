package matchstore

// ApplyLocalRequest reuses the pure actor-validated lifecycle adapter for an
// offline owner. This local capability does not authorize online results and
// never invokes Store, persistence, network services or economy settlement.
// The returned Envelope is private local storage, never presentation data.
func ApplyLocalRequest(e Envelope, seat int, r Request) (Envelope, error) {
	return applyRequest(e, seat, r)
}

// ApplyLocalAutomatic is available only to a local engine scheduler. Callers
// must never expose this capability as an actor/transport command. Generated
// chance remains private in the returned envelope for durable local recovery.
func ApplyLocalAutomatic(e Envelope, r ServerRequest) (Envelope, error) {
	return applyServerRequest(e, r)
}
