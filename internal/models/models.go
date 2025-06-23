package models

import (
	"time"

	"github.com/google/uuid"
)

// Borrower represents a loan applicant
type Borrower struct {
	ID               uuid.UUID `json:"id" db:"id"`
	FullName         string    `json:"full_name" db:"full_name"`
	Email            string    `json:"email" db:"email"`
	Age              int       `json:"age" db:"age"`
	MonthlyIncome    float64   `json:"monthly_income" db:"monthly_income"`
	EmploymentStatus string    `json:"employment_status" db:"employment_status"`
	Dependents       int       `json:"dependents" db:"dependents"`
	ExistingLoans    float64   `json:"existing_loans" db:"existing_loans"`
	SavingsBalance   float64   `json:"savings_balance" db:"savings_balance"`
	RiskScore        int       `json:"risk_score" db:"risk_score"`
	ScoreExplanation string    `json:"score_explanation" db:"score_explanation"`
	Decision         string    `json:"decision" db:"decision"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Rule represents a scoring rule
type Rule struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Field       string    `json:"field" db:"field"`
	Operator    string    `json:"operator" db:"operator"`
	Value       string    `json:"value" db:"value"`
	Weight      int       `json:"weight" db:"weight"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RuleOperator represents available rule operators
type RuleOperator string

const (
	OperatorLessThan         RuleOperator = "lt"
	OperatorLessThanEqual    RuleOperator = "lte"
	OperatorGreaterThan      RuleOperator = "gt"
	OperatorGreaterThanEqual RuleOperator = "gte"
	OperatorEqual            RuleOperator = "eq"
	OperatorNotEqual         RuleOperator = "neq"
	OperatorContains         RuleOperator = "contains"
	OperatorNotContains      RuleOperator = "not_contains"
)

// ScoringResult holds the result of scoring calculation
type ScoringResult struct {
	Score        int               `json:"score"`
	Explanation  string            `json:"explanation"`
	Decision     string            `json:"decision"`
	RulesApplied []RuleApplication `json:"rules_applied"`
}

// RuleApplication tracks which rules were applied and their impact
type RuleApplication struct {
	RuleID      uuid.UUID `json:"rule_id"`
	RuleName    string    `json:"rule_name"`
	Weight      int       `json:"weight"`
	Description string    `json:"description"`
}

// CSVUploadResult holds the result of CSV upload
type CSVUploadResult struct {
	Success       bool     `json:"success"`
	Message       string   `json:"message"`
	RecordsLoaded int      `json:"records_loaded"`
	Errors        []string `json:"errors,omitempty"`
}

// Decision thresholds
const (
	DecisionAccept = "ACCEPT"
	DecisionReview = "REVIEW"
	DecisionReject = "REJECT"
)

// GetDecision returns the decision based on score
func GetDecision(score int) string {
	switch {
	case score >= 80:
		return DecisionAccept
	case score >= 60:
		return DecisionReview
	default:
		return DecisionReject
	}
}

// BorrowerFilter holds filtering options for borrowers
type BorrowerFilter struct {
	MinScore   *int       `json:"min_score,omitempty"`
	MaxScore   *int       `json:"max_score,omitempty"`
	Decision   string     `json:"decision,omitempty"`
	RuleID     *uuid.UUID `json:"rule_id,omitempty"`
	SearchTerm string     `json:"search_term,omitempty"`
}
