package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tomassar/credit-score-evaluator/internal/models"
	"github.com/tomassar/credit-score-evaluator/internal/storage"

	"github.com/google/uuid"
)

// BorrowerService handles borrower-related operations
type BorrowerService struct {
	db *storage.DB
}

// NewBorrowerService creates a new borrower service
func NewBorrowerService(db *storage.DB) *BorrowerService {
	return &BorrowerService{db: db}
}

// CreateBorrower creates a new borrower
func (s *BorrowerService) CreateBorrower(borrower *models.Borrower) error {
	borrower.ID = uuid.New()
	borrower.CreatedAt = time.Now()
	borrower.UpdatedAt = time.Now()

	query := `
		INSERT INTO borrowers (
			id, full_name, email, age, monthly_income, employment_status,
			dependents, existing_loans, savings_balance, risk_score,
			score_explanation, decision, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		borrower.ID.String(), borrower.FullName, borrower.Email, borrower.Age,
		borrower.MonthlyIncome, borrower.EmploymentStatus, borrower.Dependents,
		borrower.ExistingLoans, borrower.SavingsBalance, borrower.RiskScore,
		borrower.ScoreExplanation, borrower.Decision,
		borrower.CreatedAt, borrower.UpdatedAt,
	)

	return err
}

// GetBorrower retrieves a borrower by ID
func (s *BorrowerService) GetBorrower(id uuid.UUID) (*models.Borrower, error) {
	borrower := &models.Borrower{}

	query := `
		SELECT id, full_name, email, age, monthly_income, employment_status,
			   dependents, existing_loans, savings_balance, risk_score,
			   score_explanation, decision, created_at, updated_at
		FROM borrowers WHERE id = ?
	`

	var idStr string
	err := s.db.QueryRow(query, id.String()).Scan(
		&idStr, &borrower.FullName, &borrower.Email, &borrower.Age,
		&borrower.MonthlyIncome, &borrower.EmploymentStatus, &borrower.Dependents,
		&borrower.ExistingLoans, &borrower.SavingsBalance, &borrower.RiskScore,
		&borrower.ScoreExplanation, &borrower.Decision,
		&borrower.CreatedAt, &borrower.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	borrower.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	return borrower, nil
}

// GetAllBorrowers retrieves all borrowers with optional filtering
func (s *BorrowerService) GetAllBorrowers(filter *models.BorrowerFilter) ([]*models.Borrower, error) {
	query := `
		SELECT id, full_name, email, age, monthly_income, employment_status,
			   dependents, existing_loans, savings_balance, risk_score,
			   score_explanation, decision, created_at, updated_at
		FROM borrowers
	`

	var args []interface{}
	var conditions []string

	if filter != nil {
		if filter.MinScore != nil {
			conditions = append(conditions, "risk_score >= ?")
			args = append(args, *filter.MinScore)
		}
		if filter.MaxScore != nil {
			conditions = append(conditions, "risk_score <= ?")
			args = append(args, *filter.MaxScore)
		}
		if filter.Decision != "" {
			conditions = append(conditions, "decision = ?")
			args = append(args, filter.Decision)
		}
		if filter.SearchTerm != "" {
			conditions = append(conditions, "(full_name LIKE ? OR email LIKE ?)")
			searchTerm := "%" + filter.SearchTerm + "%"
			args = append(args, searchTerm, searchTerm)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions, " AND ")
	}

	query += " ORDER BY updated_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var borrowers []*models.Borrower
	for rows.Next() {
		borrower := &models.Borrower{}
		var idStr string

		err := rows.Scan(
			&idStr, &borrower.FullName, &borrower.Email, &borrower.Age,
			&borrower.MonthlyIncome, &borrower.EmploymentStatus, &borrower.Dependents,
			&borrower.ExistingLoans, &borrower.SavingsBalance, &borrower.RiskScore,
			&borrower.ScoreExplanation, &borrower.Decision,
			&borrower.CreatedAt, &borrower.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		borrower.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %w", err)
		}

		borrowers = append(borrowers, borrower)
	}

	return borrowers, nil
}

// UpdateBorrower updates a borrower's information
func (s *BorrowerService) UpdateBorrower(borrower *models.Borrower) error {
	borrower.UpdatedAt = time.Now()

	query := `
		UPDATE borrowers SET
			full_name = ?, email = ?, age = ?, monthly_income = ?,
			employment_status = ?, dependents = ?, existing_loans = ?,
			savings_balance = ?, risk_score = ?, score_explanation = ?,
			decision = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		borrower.FullName, borrower.Email, borrower.Age, borrower.MonthlyIncome,
		borrower.EmploymentStatus, borrower.Dependents, borrower.ExistingLoans,
		borrower.SavingsBalance, borrower.RiskScore, borrower.ScoreExplanation,
		borrower.Decision, borrower.UpdatedAt, borrower.ID.String(),
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

// UpdateBorrowerScore updates only the risk score and related fields
func (s *BorrowerService) UpdateBorrowerScore(id uuid.UUID, score int, explanation, decision string) error {
	query := `
		UPDATE borrowers SET
			risk_score = ?, score_explanation = ?, decision = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query, score, explanation, decision, time.Now(), id.String())
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

// DeleteBorrower deletes a borrower by ID
func (s *BorrowerService) DeleteBorrower(id uuid.UUID) error {
	query := "DELETE FROM borrowers WHERE id = ?"
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

// DeleteAllBorrowers deletes all borrowers (for testing/reset)
func (s *BorrowerService) DeleteAllBorrowers() error {
	_, err := s.db.Exec("DELETE FROM borrowers")
	return err
}

// GetBorrowerCount returns the total count of borrowers
func (s *BorrowerService) GetBorrowerCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM borrowers").Scan(&count)
	return count, err
}

// joinConditions joins SQL conditions with a separator
func joinConditions(conditions []string, separator string) string {
	if len(conditions) == 0 {
		return ""
	}

	result := conditions[0]
	for i := 1; i < len(conditions); i++ {
		result += separator + conditions[i]
	}
	return result
}
