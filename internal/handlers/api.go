package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tomassar/credit-score-evaluator/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UploadCSV handles CSV file upload
func (h *Handler) UploadCSV(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, header, err := r.FormFile("csv_file")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		h.writeError(w, http.StatusBadRequest, "Only CSV files are allowed")
		return
	}

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to parse CSV file")
		return
	}

	if len(records) == 0 {
		h.writeError(w, http.StatusBadRequest, "CSV file is empty")
		return
	}

	// Process header row
	headers := records[0]
	expectedHeaders := []string{
		"full_name", "email", "age", "monthly_income", "employment_status",
		"dependents", "existing_loans", "savings_balance",
	}

	// Map headers to indices
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Validate required headers
	var missingHeaders []string
	for _, expected := range expectedHeaders {
		if _, exists := headerMap[expected]; !exists {
			missingHeaders = append(missingHeaders, expected)
		}
	}

	if len(missingHeaders) > 0 {
		h.writeError(w, http.StatusBadRequest,
			fmt.Sprintf("Missing required columns: %s", strings.Join(missingHeaders, ", ")))
		return
	}

	// Clear existing data if requested
	if r.FormValue("replace") == "true" {
		if err := h.borrowerService.DeleteAllBorrowers(); err != nil {
			h.writeError(w, http.StatusInternalServerError, "Failed to clear existing data")
			return
		}
	}

	// Process data rows
	var errors []string
	var successCount int

	for i, record := range records[1:] { // Skip header row
		borrower := &models.Borrower{}
		rowNum := i + 2 // Account for 1-based indexing and header row

		// Parse fields
		borrower.FullName = strings.TrimSpace(record[headerMap["full_name"]])
		borrower.Email = strings.TrimSpace(record[headerMap["email"]])

		// Parse numeric fields with validation
		if age, err := strconv.Atoi(strings.TrimSpace(record[headerMap["age"]])); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Invalid age", rowNum))
			continue
		} else {
			borrower.Age = age
		}

		if income, err := strconv.ParseFloat(strings.TrimSpace(record[headerMap["monthly_income"]]), 64); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Invalid monthly income", rowNum))
			continue
		} else {
			borrower.MonthlyIncome = income
		}

		borrower.EmploymentStatus = strings.TrimSpace(record[headerMap["employment_status"]])

		if deps, err := strconv.Atoi(strings.TrimSpace(record[headerMap["dependents"]])); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Invalid dependents", rowNum))
			continue
		} else {
			borrower.Dependents = deps
		}

		if loans, err := strconv.ParseFloat(strings.TrimSpace(record[headerMap["existing_loans"]]), 64); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Invalid existing loans", rowNum))
			continue
		} else {
			borrower.ExistingLoans = loans
		}

		if savings, err := strconv.ParseFloat(strings.TrimSpace(record[headerMap["savings_balance"]]), 64); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Invalid savings balance", rowNum))
			continue
		} else {
			borrower.SavingsBalance = savings
		}

		// Basic validation
		if borrower.FullName == "" || borrower.Email == "" {
			errors = append(errors, fmt.Sprintf("Row %d: Missing required fields", rowNum))
			continue
		}

		// Calculate initial risk score
		result, err := h.scoringService.CalculateScore(borrower)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Failed to calculate score", rowNum))
			continue
		}

		borrower.RiskScore = result.Score
		borrower.ScoreExplanation = result.Explanation
		borrower.Decision = result.Decision

		// Save borrower
		if err := h.borrowerService.CreateBorrower(borrower); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Failed to save borrower", rowNum))
			continue
		}

		successCount++
	}

	// Prepare response
	result := models.CSVUploadResult{
		Success:       successCount > 0,
		RecordsLoaded: successCount,
		Errors:        errors,
	}

	if successCount > 0 {
		if len(errors) > 0 {
			result.Message = fmt.Sprintf("Processed %d records with %d errors", successCount, len(errors))
		} else {
			result.Message = fmt.Sprintf("Successfully processed %d records", successCount)
		}
		h.setFlashMessage(w, r, "success", result.Message)
	} else {
		result.Message = "No records were processed successfully"
		h.setFlashMessage(w, r, "error", result.Message)
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetBorrowers returns all borrowers with optional filtering
func (h *Handler) GetBorrowers(w http.ResponseWriter, r *http.Request) {
	filter := &models.BorrowerFilter{}

	// Parse query parameters
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

	borrowers, err := h.borrowerService.GetAllBorrowers(filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get borrowers")
		return
	}

	h.writeJSON(w, http.StatusOK, borrowers)
}

// ExportBorrowers exports borrowers to CSV
func (h *Handler) ExportBorrowers(w http.ResponseWriter, r *http.Request) {
	borrowers, err := h.borrowerService.GetAllBorrowers(nil)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get borrowers")
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=risk_scores_%s.csv",
		time.Now().Format("2006-01-02_15-04-05")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	headers := []string{
		"Full Name", "Email", "Age", "Monthly Income", "Employment Status",
		"Dependents", "Existing Loans", "Savings Balance", "Risk Score",
		"Score Explanation", "Decision",
	}
	writer.Write(headers)

	// Write data
	for _, borrower := range borrowers {
		record := []string{
			borrower.FullName,
			borrower.Email,
			strconv.Itoa(borrower.Age),
			fmt.Sprintf("%.2f", borrower.MonthlyIncome),
			borrower.EmploymentStatus,
			strconv.Itoa(borrower.Dependents),
			fmt.Sprintf("%.2f", borrower.ExistingLoans),
			fmt.Sprintf("%.2f", borrower.SavingsBalance),
			strconv.Itoa(borrower.RiskScore),
			borrower.ScoreExplanation,
			borrower.Decision,
		}
		writer.Write(record)
	}
}

// GetRules returns all rules
func (h *Handler) GetRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get rules")
		return
	}

	h.writeJSON(w, http.StatusOK, rules)
}

// CreateRule creates a new rule
func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate rule
	if err := h.scoringService.ValidateRule(&rule); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create rule
	if err := h.ruleService.CreateRule(&rule); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create rule")
		return
	}

	// Recalculate all scores
	if err := h.scoringService.RecalculateAllScores(h.borrowerService); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to recalculate scores")
		return
	}

	h.setFlashMessage(w, r, "success", "Rule created successfully")
	h.writeJSON(w, http.StatusCreated, rule)
}

// UpdateRule updates an existing rule
func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid rule ID")
		return
	}

	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	rule.ID = id

	// Validate rule
	if err := h.scoringService.ValidateRule(&rule); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update rule
	if err := h.ruleService.UpdateRule(&rule); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update rule")
		return
	}

	// Recalculate all scores
	if err := h.scoringService.RecalculateAllScores(h.borrowerService); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to recalculate scores")
		return
	}

	h.setFlashMessage(w, r, "success", "Rule updated successfully")
	h.writeJSON(w, http.StatusOK, rule)
}

// DeleteRule deletes a rule
func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid rule ID")
		return
	}

	if err := h.ruleService.DeleteRule(id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete rule")
		return
	}

	// Recalculate all scores
	if err := h.scoringService.RecalculateAllScores(h.borrowerService); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to recalculate scores")
		return
	}

	h.setFlashMessage(w, r, "success", "Rule deleted successfully")
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Rule deleted successfully"})
}

// RecalculateScores manually recalculates all scores
func (h *Handler) RecalculateScores(w http.ResponseWriter, r *http.Request) {
	if err := h.scoringService.RecalculateAllScores(h.borrowerService); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to recalculate scores")
		return
	}

	h.setFlashMessage(w, r, "success", "All scores recalculated successfully")
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Scores recalculated successfully"})
}
