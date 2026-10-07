package transport

import (
	"errors"
	"net/http"

	"github.com/metaphy6/cgms/backend/internal/economy"
)

func (s *Server) economyRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /v1/economy", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		if s.economy == nil {
			s.write(w, 200, map[string]bool{"enabled": false})
			return
		}
		v, err := s.economy.Account(r.Context(), a.ID)
		if err != nil {
			s.economyError(w, err)
			return
		}
		s.write(w, 200, v)
	})
	m.HandleFunc("POST /v1/economy/passes", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		if s.economy == nil {
			s.fail(w, 409, "ECONOMY_DISABLED")
			return
		}
		var body struct {
			OperationID string `json:"operation_id"`
			Days        int    `json:"days"`
		}
		if decode(r, &body) != nil {
			s.fail(w, 400, "INVALID_REQUEST")
			return
		}
		v, err := s.economy.BuyPass(r.Context(), a.ID, body.OperationID, body.Days)
		if err != nil {
			s.economyError(w, err)
			return
		}
		s.wakeAutomatic()
		s.write(w, 200, v)
	})
	m.HandleFunc("GET /v1/economy/passes/{operation}", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		if s.economy == nil {
			s.fail(w, 409, "ECONOMY_DISABLED")
			return
		}
		v, err := s.economy.LookupPass(r.Context(), a.ID, r.PathValue("operation"))
		if err != nil {
			s.economyError(w, err)
			return
		}
		s.write(w, 200, v)
	})
}

func (s *Server) economyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, economy.ErrProductUnavailable):
		s.fail(w, 409, "PRODUCT_UNAVAILABLE")
	case errors.Is(err, economy.ErrInvalid):
		s.fail(w, 400, "INVALID_REQUEST")
	case errors.Is(err, economy.ErrIneligible):
		s.fail(w, 403, "FORBIDDEN")
	case errors.Is(err, economy.ErrNotFound):
		s.fail(w, 404, "OPERATION_NOT_FOUND")
	case errors.Is(err, economy.ErrConflict):
		s.fail(w, 409, "ECONOMY_CONFLICT")
	case errors.Is(err, economy.ErrFunds):
		s.fail(w, 409, "INSUFFICIENT_DIRT")
	case errors.Is(err, economy.ErrAllowance):
		s.fail(w, 409, "ALLOWANCE_EXHAUSTED")
	case errors.Is(err, economy.ErrOutcomeUnknown):
		s.fail(w, 503, "OUTCOME_UNKNOWN")
	default:
		s.err(w, err)
	}
}
