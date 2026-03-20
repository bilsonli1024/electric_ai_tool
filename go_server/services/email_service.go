package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"electric_ai_tool/go_server/config"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) GenerateCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code
}

func (s *EmailService) SaveVerificationCode(email, code, purpose string) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	_, err := db.Exec(
		`INSERT INTO email_verification_codes_tab (email, code, purpose, expires_at) 
		 VALUES (?, ?, ?, ?)`,
		email, code, purpose, expiresAt,
	)
	return err
}

func (s *EmailService) VerifyCode(email, code, purpose string) (bool, error) {
	db := config.GetDB()
	if db == nil {
		return false, fmt.Errorf("database not initialized")
	}

	var id int64
	var used int
	var expiresAt time.Time

	err := db.QueryRow(
		`SELECT id, used, expires_at FROM email_verification_codes_tab 
		 WHERE email = ? AND code = ? AND purpose = ? 
		 ORDER BY created_at DESC LIMIT 1`,
		email, code, purpose,
	).Scan(&id, &used, &expiresAt)

	if err != nil {
		return false, err
	}

	if used == 1 {
		return false, fmt.Errorf("verification code already used")
	}

	if time.Now().After(expiresAt) {
		return false, fmt.Errorf("verification code expired")
	}

	_, err = db.Exec(
		`UPDATE email_verification_codes_tab SET used = 1 WHERE id = ?`,
		id,
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *EmailService) SendVerificationCode(email, code, purpose string) error {
	fmt.Printf("📧 Send verification code to %s: %s (purpose: %s)\n", email, code, purpose)
	return nil
}

func (s *EmailService) SendTestEmail(email, code string) error {
	fmt.Printf("🧪 TEST EMAIL - To: %s, Code: %s\n", email, code)
	fmt.Printf("==========================================\n")
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Verification Code: %s\n", code)
	fmt.Printf("==========================================\n")
	return nil
}
