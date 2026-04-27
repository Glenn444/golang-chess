package emails

import (
	"fmt"

	"github.com/resend/resend-go/v3"
)

type EmailClient struct{
    ApiKey string
}



func NewEmailClient(apiKey string)*EmailClient{
    return &EmailClient{ApiKey: apiKey}
}


func (c *EmailClient) SendEmailOTP(to, otp string) error {
    client := resend.NewClient(c.ApiKey)

    html := fmt.Sprintf(`
    <div style="font-family: Arial, sans-serif; max-width: 520px; margin: 0 auto;">
      <div style="background: #1a1a2e; padding: 20px 24px; border-radius: 8px 8px 0 0;">
        <span style="color: white; font-weight: bold; font-size: 16px;">AccountYetu</span>
      </div>
      <div style="border: 1px solid #e5e7eb; border-top: none; padding: 32px 24px; border-radius: 0 0 8px 8px;">
        <p style="font-size: 13px; color: #6b7280; margin: 0 0 24px;">Verify your email address</p>
        <p style="font-size: 15px; font-weight: 600; margin: 0 0 8px; color: #111827;">Your verification code</p>
        <p style="font-size: 14px; color: #6b7280; margin: 0 0 24px; line-height: 1.6;">
          Use the code below to verify your email address. 
          This code expires in <strong style="color: #111827;">15 minutes</strong>.
        </p>
        <div style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 8px; padding: 20px;
                    text-align: center; letter-spacing: 8px; font-size: 28px; font-weight: 600;
                    font-family: monospace; color: #111827; margin: 0 0 24px;">
          %s
        </div>
        <p style="font-size: 13px; color: #6b7280; line-height: 1.6; margin: 0 0 24px;">
          If you didn't request this, you can safely ignore this email.
          Someone may have typed your email address by mistake.
        </p>
        <div style="border-top: 1px solid #e5e7eb; padding-top: 20px;">
          <p style="font-size: 12px; color: #9ca3af; margin: 0;">
            This is an automated message from AccountYetu. Please do not reply to this email.
          </p>
        </div>
      </div>
    </div>`, otp)

    params := &resend.SendEmailRequest{
        From:    "AccountYetu <info@accountyetu.com>",
        To:      []string{to},
        Subject: "Your Chess Game verification code",
        Html:    html,
    }

    sent, err := client.Emails.Send(params)
    if err != nil {
        return fmt.Errorf("failed to send OTP email: %w", err)
    }

    fmt.Println("email sent:", sent.Id)
    return nil
}