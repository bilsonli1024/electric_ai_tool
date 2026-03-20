package services

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"
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

func (s *EmailService) sendEmailViaSMTP(to, subject, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM")

	// 如果没有配置SMTP，只打印日志
	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Printf("⚠️  SMTP not configured, email not sent to %s", to)
		log.Printf("📧 Subject: %s", subject)
		log.Printf("📧 Body: %s", body)
		return fmt.Errorf("SMTP not configured in .env file")
	}

	if fromEmail == "" {
		fromEmail = smtpUser
	}

	if smtpPort == "" {
		smtpPort = "587"
	}

	// 构建邮件内容
	message := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		fromEmail, to, subject, body,
	))

	// 设置认证
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// 发送邮件
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, fromEmail, []string{to}, message)
	
	if err != nil {
		log.Printf("✗ Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("✓ Email sent successfully to %s", to)
	return nil
}

func (s *EmailService) SendVerificationCode(email, code, purpose string) error {
	purposeText := map[string]string{
		"register": "注册账号",
		"reset":    "重置密码",
	}[purpose]

	if purposeText == "" {
		purposeText = "验证"
	}

	subject := fmt.Sprintf("【Electric AI】您的%s验证码", purposeText)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 10px 10px; }
        .code { background: white; border: 2px dashed #667eea; padding: 20px; text-align: center; margin: 20px 0; border-radius: 10px; }
        .code-number { font-size: 32px; font-weight: bold; color: #667eea; letter-spacing: 5px; }
        .footer { text-align: center; color: #999; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Electric AI Tool</h1>
            <p>智能营销工具平台</p>
        </div>
        <div class="content">
            <h2>您好！</h2>
            <p>您正在进行<strong>%s</strong>操作，请使用以下验证码完成验证：</p>
            <div class="code">
                <div class="code-number">%s</div>
            </div>
            <p><strong>验证码有效期为10分钟</strong>，请尽快使用。</p>
            <p>如果这不是您本人的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>&copy; 2026 Electric AI Tool. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, purposeText, code)

	return s.sendEmailViaSMTP(email, subject, body)
}

func (s *EmailService) SendTestEmail(email, code string) error {
	log.Printf("🧪 TEST EMAIL - To: %s, Code: %s", email, code)
	log.Printf("==========================================")
	log.Printf("Email: %s", email)
	log.Printf("Verification Code: %s", code)
	log.Printf("==========================================")

	// 测试时也尝试真正发送
	subject := "【Electric AI】测试验证码"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #ff6b6b 0%%, #ee5a6f 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 10px 10px; }
        .code { background: white; border: 2px dashed #ff6b6b; padding: 20px; text-align: center; margin: 20px 0; border-radius: 10px; }
        .code-number { font-size: 32px; font-weight: bold; color: #ff6b6b; letter-spacing: 5px; }
        .footer { text-align: center; color: #999; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 测试邮件</h1>
            <p>Electric AI Tool - 邮件发送测试</p>
        </div>
        <div class="content">
            <h2>这是一封测试邮件</h2>
            <p>如果您收到这封邮件，说明邮件服务配置成功！</p>
            <div class="code">
                <div class="code-number">%s</div>
            </div>
            <p>测试验证码（仅供测试，无实际用途）</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>&copy; 2026 Electric AI Tool. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, code)

	return s.sendEmailViaSMTP(email, subject, body)
}

