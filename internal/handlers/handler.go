package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tomassar/credit-score-evaluator/internal/services"

	"github.com/gorilla/sessions"
)

// Handler holds all the dependencies for HTTP handlers
type Handler struct {
	borrowerService *services.BorrowerService
	ruleService     *services.RuleService
	scoringService  *services.ScoringService
	sessionStore    *sessions.CookieStore
}

// New creates a new handler instance
func New(
	borrowerService *services.BorrowerService,
	ruleService *services.RuleService,
	scoringService *services.ScoringService,
	sessionStore *sessions.CookieStore,
) *Handler {
	return &Handler{
		borrowerService: borrowerService,
		ruleService:     ruleService,
		scoringService:  scoringService,
		sessionStore:    sessionStore,
	}
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

// getSession gets the session for the request
func (h *Handler) getSession(r *http.Request) (*sessions.Session, error) {
	return h.sessionStore.Get(r, "risk-calculator-session")
}

// setFlashMessage sets a flash message in the session
func (h *Handler) setFlashMessage(w http.ResponseWriter, r *http.Request, msgType, message string) {
	session, err := h.getSession(r)
	if err != nil {
		return
	}

	session.AddFlash(map[string]string{
		"type":    msgType,
		"message": message,
	})

	session.Save(r, w)
}

// getFlashMessages retrieves and clears flash messages from the session
func (h *Handler) getFlashMessages(w http.ResponseWriter, r *http.Request) []map[string]string {
	session, err := h.getSession(r)
	if err != nil {
		return nil
	}

	flashes := session.Flashes()
	if len(flashes) == 0 {
		return nil
	}

	session.Save(r, w)

	var messages []map[string]string
	for _, flash := range flashes {
		if flashMap, ok := flash.(map[string]string); ok {
			messages = append(messages, flashMap)
		}
	}

	return messages
}
