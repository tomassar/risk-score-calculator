package handlers

import (
	"net/http"

	"github.com/tomassar/credit-score-evaluator/web/templates"
)

// HomePage renders the home page
func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	// Get statistics
	borrowerCount, _ := h.borrowerService.GetBorrowerCount()
	allRules, _ := h.ruleService.GetAllRules()
	activeRuleCount := 0
	for _, rule := range allRules {
		if rule.IsActive {
			activeRuleCount++
		}
	}

	// Get flash messages
	flashes := h.getFlashMessages(w, r)

	data := templates.HomeData{
		BorrowerCount:   borrowerCount,
		ActiveRuleCount: activeRuleCount,
		TotalRuleCount:  len(allRules),
		FlashMessages:   flashes,
	}

	component := templates.HomePage(data)
	component.Render(r.Context(), w)
}

// UploadPage renders the upload page
func (h *Handler) UploadPage(w http.ResponseWriter, r *http.Request) {
	flashes := h.getFlashMessages(w, r)

	data := templates.UploadData{
		FlashMessages: flashes,
	}

	component := templates.UploadPage(data)
	component.Render(r.Context(), w)
}

// BorrowersPage renders the borrowers page
func (h *Handler) BorrowersPage(w http.ResponseWriter, r *http.Request) {
	// Get borrowers
	borrowers, err := h.borrowerService.GetAllBorrowers(nil)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to load borrowers")
		return
	}

	// Get flash messages
	flashes := h.getFlashMessages(w, r)

	data := templates.BorrowersData{
		Borrowers:     borrowers,
		FlashMessages: flashes,
	}

	component := templates.BorrowersPage(data)
	component.Render(r.Context(), w)
}

// RulesPage renders the rules page
func (h *Handler) RulesPage(w http.ResponseWriter, r *http.Request) {
	// Get rules
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to load rules")
		return
	}

	// Get flash messages
	flashes := h.getFlashMessages(w, r)

	data := templates.RulesData{
		Rules:         rules,
		FlashMessages: flashes,
	}

	component := templates.RulesPage(data)
	component.Render(r.Context(), w)
}
