package handlers

import (
	"net/http"
	"strconv"

	"github.com/tomassar/credit-score-evaluator/internal/models"
	"github.com/tomassar/credit-score-evaluator/web/templates"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// BorrowersTable renders the borrowers table partial
func (h *Handler) BorrowersTable(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	filter := &models.BorrowerFilter{}

	if minScoreStr := r.URL.Query().Get("min_score"); minScoreStr != "" {
		if minScore, err := strconv.Atoi(minScoreStr); err == nil {
			filter.MinScore = &minScore
		}
	}

	if maxScoreStr := r.URL.Query().Get("max_score"); maxScoreStr != "" {
		if maxScore, err := strconv.Atoi(maxScoreStr); err == nil {
			filter.MaxScore = &maxScore
		}
	}

	if decision := r.URL.Query().Get("decision"); decision != "" {
		filter.Decision = decision
	}

	if searchTerm := r.URL.Query().Get("search"); searchTerm != "" {
		filter.SearchTerm = searchTerm
	}

	// Get filtered borrowers
	borrowers, err := h.borrowerService.GetAllBorrowers(filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to load borrowers")
		return
	}

	component := templates.BorrowersTable(borrowers)
	component.Render(r.Context(), w)
}

// RulesList renders the rules list partial
func (h *Handler) RulesList(w http.ResponseWriter, r *http.Request) {
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to load rules")
		return
	}

	component := templates.RulesList(rules)
	component.Render(r.Context(), w)
}

// UploadForm renders the upload form partial
func (h *Handler) UploadForm(w http.ResponseWriter, r *http.Request) {
	component := templates.UploadForm()
	component.Render(r.Context(), w)
}

// RuleModal renders the rule modal for create/edit
func (h *Handler) RuleModal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	var rule *models.Rule
	isEditing := false

	// If ID is provided, it's an edit modal
	if idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid rule ID")
			return
		}

		rule, err = h.ruleService.GetRule(id)
		if err != nil {
			h.writeError(w, http.StatusNotFound, "Rule not found")
			return
		}
		isEditing = true
	} else {
		// Create new rule
		rule = &models.Rule{
			IsActive: true,
		}
	}

	component := templates.RuleModal(rule, isEditing)
	component.Render(r.Context(), w)
}

// ProcessRule handles rule creation/update via HTMX
func (h *Handler) ProcessRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	var rule models.Rule
	var isEditing bool

	// Parse form data
	rule.Name = r.FormValue("name")
	rule.Description = r.FormValue("description")
	rule.Field = r.FormValue("field")
	rule.Operator = r.FormValue("operator")
	rule.Value = r.FormValue("value")

	weight, err := strconv.Atoi(r.FormValue("weight"))
	if err != nil {
		h.renderFlashMessage(w, r, "error", "Invalid weight value")
		return
	}
	rule.Weight = weight
	rule.IsActive = r.FormValue("is_active") == "on"

	// Validate rule
	if err := h.scoringService.ValidateRule(&rule); err != nil {
		h.renderFlashMessage(w, r, "error", err.Error())
		return
	}

	// Create or update
	if idStr != "" {
		// Update existing rule
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.renderFlashMessage(w, r, "error", "Invalid rule ID")
			return
		}
		rule.ID = id
		isEditing = true

		if err := h.ruleService.UpdateRule(&rule); err != nil {
			h.renderFlashMessage(w, r, "error", "Failed to update rule")
			return
		}
	} else {
		// Create new rule
		if err := h.ruleService.CreateRule(&rule); err != nil {
			h.renderFlashMessage(w, r, "error", "Failed to create rule")
			return
		}
	}

	// Recalculate scores
	if err := h.scoringService.RecalculateAllScores(h.borrowerService); err != nil {
		h.renderFlashMessage(w, r, "error", "Rule saved but failed to recalculate scores")
		return
	}

	// Return success response with updated rules list
	w.Header().Set("HX-Trigger", "rule-saved,close-modal")

	action := "created"
	if isEditing {
		action = "updated"
	}

	h.renderFlashMessage(w, r, "success", "Rule "+action+" successfully")
}

// DeleteRuleHTMX handles rule deletion via HTMX
func (h *Handler) DeleteRuleHTMX(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.renderFlashMessage(w, r, "error", "Invalid rule ID")
		return
	}

	if err := h.ruleService.DeleteRule(id); err != nil {
		h.renderFlashMessage(w, r, "error", "Failed to delete rule")
		return
	}

	// Recalculate scores
	if err := h.scoringService.RecalculateAllScores(h.borrowerService); err != nil {
		h.renderFlashMessage(w, r, "error", "Rule deleted but failed to recalculate scores")
		return
	}

	// Return updated rules list
	w.Header().Set("HX-Trigger", "rule-deleted")
	h.renderFlashMessage(w, r, "success", "Rule deleted successfully")
}

// FlashMessage renders a flash message component
func (h *Handler) FlashMessage(w http.ResponseWriter, r *http.Request) {
	msgType := r.URL.Query().Get("type")
	message := r.URL.Query().Get("message")

	if msgType == "" || message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	component := templates.FlashMessage(msgType, message)
	component.Render(r.Context(), w)
}

// renderFlashMessage is a helper to render flash messages
func (h *Handler) renderFlashMessage(w http.ResponseWriter, r *http.Request, msgType, message string) {
	w.Header().Set("HX-Trigger", "show-flash")
	w.Header().Set("HX-Trigger-After-Settle", "show-flash")

	component := templates.FlashMessage(msgType, message)
	component.Render(r.Context(), w)
}

// DashboardStats renders updated dashboard statistics
func (h *Handler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	// Get statistics
	borrowerCount, _ := h.borrowerService.GetBorrowerCount()
	allRules, _ := h.ruleService.GetAllRules()
	activeRuleCount := 0
	for _, rule := range allRules {
		if rule.IsActive {
			activeRuleCount++
		}
	}

	data := struct {
		BorrowerCount   int
		ActiveRuleCount int
		TotalRuleCount  int
	}{
		BorrowerCount:   borrowerCount,
		ActiveRuleCount: activeRuleCount,
		TotalRuleCount:  len(allRules),
	}

	component := templates.DashboardStats(data)
	component.Render(r.Context(), w)
}

// RecalculateScoresHTMX handles score recalculation via HTMX
func (h *Handler) RecalculateScoresHTMX(w http.ResponseWriter, r *http.Request) {
	if err := h.scoringService.RecalculateAllScores(h.borrowerService); err != nil {
		h.renderFlashMessage(w, r, "error", "Failed to recalculate scores")
		return
	}

	// Trigger refresh of relevant components
	w.Header().Set("HX-Trigger", "scores-recalculated")
	h.renderFlashMessage(w, r, "success", "All scores recalculated successfully")
}

// UploadStatus renders upload status after CSV processing
func (h *Handler) UploadStatus(w http.ResponseWriter, r *http.Request) {
	success := r.URL.Query().Get("success") == "true"
	message := r.URL.Query().Get("message")
	recordsLoaded := r.URL.Query().Get("records_loaded")

	if message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	component := templates.UploadStatus(success, message, recordsLoaded)
	component.Render(r.Context(), w)
}
