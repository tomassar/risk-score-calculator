package services

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/tomassar/credit-score-evaluator/internal/models"
)

// ScoringService handles risk score calculations
type ScoringService struct {
	ruleService *RuleService
}

// NewScoringService creates a new scoring service
func NewScoringService(ruleService *RuleService) *ScoringService {
	return &ScoringService{
		ruleService: ruleService,
	}
}

// CalculateScore calculates the risk score for a borrower based on active rules
func (s *ScoringService) CalculateScore(borrower *models.Borrower) (*models.ScoringResult, error) {
	// Get active rules
	rules, err := s.ruleService.GetActiveRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get active rules: %w", err)
	}

	// Start with base score
	baseScore := 100
	finalScore := baseScore

	var rulesApplied []models.RuleApplication
	var explanations []string

	// Apply each rule
	for _, rule := range rules {
		applies, err := s.evaluateRule(borrower, rule)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate rule %s: %w", rule.Name, err)
		}

		if applies {
			finalScore += rule.Weight

			ruleApp := models.RuleApplication{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Weight:      rule.Weight,
				Description: rule.Description,
			}
			rulesApplied = append(rulesApplied, ruleApp)

			// Create explanation
			weightStr := ""
			if rule.Weight > 0 {
				weightStr = fmt.Sprintf("+%d", rule.Weight)
			} else {
				weightStr = fmt.Sprintf("%d", rule.Weight)
			}
			explanations = append(explanations, fmt.Sprintf("%s: %s", weightStr, rule.Name))
		}
	}

	// Ensure score is within valid range (0-150)
	if finalScore < 0 {
		finalScore = 0
	} else if finalScore > 150 {
		finalScore = 150
	}

	// Generate explanation
	explanation := fmt.Sprintf("Base: %d", baseScore)
	if len(explanations) > 0 {
		explanation += " | " + strings.Join(explanations, " | ")
	}
	explanation += fmt.Sprintf(" = %d", finalScore)

	// Determine decision
	decision := models.GetDecision(finalScore)

	return &models.ScoringResult{
		Score:        finalScore,
		Explanation:  explanation,
		Decision:     decision,
		RulesApplied: rulesApplied,
	}, nil
}

// evaluateRule checks if a rule applies to a borrower
func (s *ScoringService) evaluateRule(borrower *models.Borrower, rule *models.Rule) (bool, error) {
	// Get the field value from the borrower
	fieldValue, err := s.getFieldValue(borrower, rule.Field)
	if err != nil {
		return false, fmt.Errorf("failed to get field value for %s: %w", rule.Field, err)
	}

	// Compare based on operator
	return s.compareValues(fieldValue, rule.Operator, rule.Value)
}

// getFieldValue extracts the field value from a borrower struct
func (s *ScoringService) getFieldValue(borrower *models.Borrower, fieldName string) (interface{}, error) {
	v := reflect.ValueOf(borrower).Elem()
	t := reflect.TypeOf(borrower).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		// Extract the field name from the json tag
		jsonName := strings.Split(jsonTag, ",")[0]

		if jsonName == fieldName {
			return v.Field(i).Interface(), nil
		}
	}

	return nil, fmt.Errorf("field %s not found", fieldName)
}

// compareValues compares two values based on the given operator
func (s *ScoringService) compareValues(fieldValue interface{}, operator, ruleValue string) (bool, error) {
	switch operator {
	case "eq":
		return s.compareEqual(fieldValue, ruleValue)
	case "neq":
		result, err := s.compareEqual(fieldValue, ruleValue)
		return !result, err
	case "lt":
		return s.compareNumeric(fieldValue, ruleValue, func(a, b float64) bool { return a < b })
	case "lte":
		return s.compareNumeric(fieldValue, ruleValue, func(a, b float64) bool { return a <= b })
	case "gt":
		return s.compareNumeric(fieldValue, ruleValue, func(a, b float64) bool { return a > b })
	case "gte":
		return s.compareNumeric(fieldValue, ruleValue, func(a, b float64) bool { return a >= b })
	case "contains":
		return s.compareContains(fieldValue, ruleValue, true)
	case "not_contains":
		return s.compareContains(fieldValue, ruleValue, false)
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// compareEqual compares for equality
func (s *ScoringService) compareEqual(fieldValue interface{}, ruleValue string) (bool, error) {
	switch v := fieldValue.(type) {
	case string:
		return strings.EqualFold(v, ruleValue), nil
	case int:
		ruleInt, err := strconv.Atoi(ruleValue)
		if err != nil {
			return false, fmt.Errorf("invalid integer rule value: %s", ruleValue)
		}
		return v == ruleInt, nil
	case float64:
		ruleFloat, err := strconv.ParseFloat(ruleValue, 64)
		if err != nil {
			return false, fmt.Errorf("invalid float rule value: %s", ruleValue)
		}
		return v == ruleFloat, nil
	case bool:
		ruleBool, err := strconv.ParseBool(ruleValue)
		if err != nil {
			return false, fmt.Errorf("invalid boolean rule value: %s", ruleValue)
		}
		return v == ruleBool, nil
	default:
		return false, fmt.Errorf("unsupported field type for equality comparison: %T", fieldValue)
	}
}

// compareNumeric compares numeric values
func (s *ScoringService) compareNumeric(fieldValue interface{}, ruleValue string, compareFn func(float64, float64) bool) (bool, error) {
	ruleFloat, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return false, fmt.Errorf("invalid numeric rule value: %s", ruleValue)
	}

	var fieldFloat float64
	switch v := fieldValue.(type) {
	case int:
		fieldFloat = float64(v)
	case float64:
		fieldFloat = v
	default:
		return false, fmt.Errorf("field is not numeric: %T", fieldValue)
	}

	return compareFn(fieldFloat, ruleFloat), nil
}

// compareContains checks if a string field contains the rule value
func (s *ScoringService) compareContains(fieldValue interface{}, ruleValue string, shouldContain bool) (bool, error) {
	fieldStr, ok := fieldValue.(string)
	if !ok {
		return false, fmt.Errorf("field is not a string: %T", fieldValue)
	}

	contains := strings.Contains(strings.ToLower(fieldStr), strings.ToLower(ruleValue))
	return contains == shouldContain, nil
}

// RecalculateAllScores recalculates scores for all borrowers
func (s *ScoringService) RecalculateAllScores(borrowerService *BorrowerService) error {
	// Get all borrowers
	borrowers, err := borrowerService.GetAllBorrowers(nil)
	if err != nil {
		return fmt.Errorf("failed to get borrowers: %w", err)
	}

	// Recalculate score for each borrower
	for _, borrower := range borrowers {
		result, err := s.CalculateScore(borrower)
		if err != nil {
			return fmt.Errorf("failed to calculate score for borrower %s: %w", borrower.ID, err)
		}

		// Update borrower with new score
		err = borrowerService.UpdateBorrowerScore(
			borrower.ID,
			result.Score,
			result.Explanation,
			result.Decision,
		)
		if err != nil {
			return fmt.Errorf("failed to update borrower score: %w", err)
		}
	}

	return nil
}

// ValidateRule validates that a rule is correctly configured
func (s *ScoringService) ValidateRule(rule *models.Rule) error {
	// Check if field is valid
	validFields := []string{
		"full_name", "email", "age", "monthly_income", "employment_status",
		"dependents", "existing_loans", "savings_balance",
	}

	fieldValid := false
	for _, validField := range validFields {
		if rule.Field == validField {
			fieldValid = true
			break
		}
	}
	if !fieldValid {
		return fmt.Errorf("invalid field: %s", rule.Field)
	}

	// Check if operator is valid
	validOperators := []string{"eq", "neq", "lt", "lte", "gt", "gte", "contains", "not_contains"}
	operatorValid := false
	for _, validOp := range validOperators {
		if rule.Operator == validOp {
			operatorValid = true
			break
		}
	}
	if !operatorValid {
		return fmt.Errorf("invalid operator: %s", rule.Operator)
	}

	// Validate value based on field type
	switch rule.Field {
	case "age", "dependents":
		if _, err := strconv.Atoi(rule.Value); err != nil {
			return fmt.Errorf("invalid integer value for field %s: %s", rule.Field, rule.Value)
		}
	case "monthly_income", "existing_loans", "savings_balance":
		if _, err := strconv.ParseFloat(rule.Value, 64); err != nil {
			return fmt.Errorf("invalid numeric value for field %s: %s", rule.Field, rule.Value)
		}
	case "full_name", "email", "employment_status":
		// String fields - no additional validation needed
	}

	// Validate weight is reasonable
	if rule.Weight < -100 || rule.Weight > 100 {
		return fmt.Errorf("weight should be between -100 and 100, got: %d", rule.Weight)
	}

	return nil
}
