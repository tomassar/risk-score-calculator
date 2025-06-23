package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tomassar/credit-score-evaluator/internal/models"
	"github.com/tomassar/credit-score-evaluator/internal/storage"

	"github.com/google/uuid"
)

// RuleService handles rule-related operations
type RuleService struct {
	db *storage.DB
}

// NewRuleService creates a new rule service
func NewRuleService(db *storage.DB) *RuleService {
	return &RuleService{db: db}
}

// CreateRule creates a new scoring rule
func (s *RuleService) CreateRule(rule *models.Rule) error {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	query := `
		INSERT INTO rules (
			id, name, description, field, operator, value, weight, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		rule.ID.String(), rule.Name, rule.Description, rule.Field,
		rule.Operator, rule.Value, rule.Weight, rule.IsActive,
		rule.CreatedAt, rule.UpdatedAt,
	)

	return err
}

// GetRule retrieves a rule by ID
func (s *RuleService) GetRule(id uuid.UUID) (*models.Rule, error) {
	rule := &models.Rule{}

	query := `
		SELECT id, name, description, field, operator, value, weight, is_active,
			   created_at, updated_at
		FROM rules WHERE id = ?
	`

	var idStr string
	err := s.db.QueryRow(query, id.String()).Scan(
		&idStr, &rule.Name, &rule.Description, &rule.Field,
		&rule.Operator, &rule.Value, &rule.Weight, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	rule.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	return rule, nil
}

// GetAllRules retrieves all rules
func (s *RuleService) GetAllRules() ([]*models.Rule, error) {
	query := `
		SELECT id, name, description, field, operator, value, weight, is_active,
			   created_at, updated_at
		FROM rules
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.Rule
	for rows.Next() {
		rule := &models.Rule{}
		var idStr string

		err := rows.Scan(
			&idStr, &rule.Name, &rule.Description, &rule.Field,
			&rule.Operator, &rule.Value, &rule.Weight, &rule.IsActive,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		rule.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// GetActiveRules retrieves only active rules
func (s *RuleService) GetActiveRules() ([]*models.Rule, error) {
	query := `
		SELECT id, name, description, field, operator, value, weight, is_active,
			   created_at, updated_at
		FROM rules
		WHERE is_active = true
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.Rule
	for rows.Next() {
		rule := &models.Rule{}
		var idStr string

		err := rows.Scan(
			&idStr, &rule.Name, &rule.Description, &rule.Field,
			&rule.Operator, &rule.Value, &rule.Weight, &rule.IsActive,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		rule.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// UpdateRule updates a rule
func (s *RuleService) UpdateRule(rule *models.Rule) error {
	rule.UpdatedAt = time.Now()

	query := `
		UPDATE rules SET
			name = ?, description = ?, field = ?, operator = ?, value = ?,
			weight = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		rule.Name, rule.Description, rule.Field, rule.Operator,
		rule.Value, rule.Weight, rule.IsActive, rule.UpdatedAt,
		rule.ID.String(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteRule deletes a rule by ID
func (s *RuleService) DeleteRule(id uuid.UUID) error {
	query := "DELETE FROM rules WHERE id = ?"
	result, err := s.db.Exec(query, id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ToggleRule toggles the active status of a rule
func (s *RuleService) ToggleRule(id uuid.UUID) error {
	query := `
		UPDATE rules SET
			is_active = NOT is_active,
			updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query, time.Now(), id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// CreateDefaultRules creates a set of default rules for initial setup
func (s *RuleService) CreateDefaultRules() error {
	defaultRules := []*models.Rule{
		{
			Name:        "Low Income Penalty",
			Description: "Penalize borrowers with monthly income below $30,000",
			Field:       "monthly_income",
			Operator:    "lt",
			Value:       "30000",
			Weight:      -20,
			IsActive:    true,
		},
		{
			Name:        "High Income Bonus",
			Description: "Bonus for borrowers with monthly income above $80,000",
			Field:       "monthly_income",
			Operator:    "gt",
			Value:       "80000",
			Weight:      15,
			IsActive:    true,
		},
		{
			Name:        "Unemployed Penalty",
			Description: "High penalty for unemployed borrowers",
			Field:       "employment_status",
			Operator:    "eq",
			Value:       "unemployed",
			Weight:      -35,
			IsActive:    true,
		},
		{
			Name:        "High Savings Bonus",
			Description: "Bonus for borrowers with savings above $50,000",
			Field:       "savings_balance",
			Operator:    "gt",
			Value:       "50000",
			Weight:      20,
			IsActive:    true,
		},
		{
			Name:        "Age Risk Factor",
			Description: "Penalty for borrowers under 21 or over 65",
			Field:       "age",
			Operator:    "lt",
			Value:       "21",
			Weight:      -15,
			IsActive:    true,
		},
		{
			Name:        "High Existing Loans",
			Description: "Penalty for borrowers with existing loans above $100,000",
			Field:       "existing_loans",
			Operator:    "gt",
			Value:       "100000",
			Weight:      -25,
			IsActive:    true,
		},
	}

	for _, rule := range defaultRules {
		if err := s.CreateRule(rule); err != nil {
			return fmt.Errorf("failed to create default rule %s: %w", rule.Name, err)
		}
	}

	return nil
}
