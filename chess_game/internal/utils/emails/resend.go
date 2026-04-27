package emails

import (
	"fmt"

	"github.com/Glenn444/golang-chess/config"
	"github.com/resend/resend-go/v3"
)



func SendEmailOTP(email string,otp string) {
	resendKey := config.Config.RESEND_API_KEY
    client := resend.NewClient("re_xxxxxxxxx")

    params := &resend.SendEmailRequest{
        From:    "Acme <onboarding@resend.dev>",
        To:      []string{"delivered@resend.dev"},
        Html:    "<strong>hello world</strong>",
        Subject: "Hello from Golang",
        Cc:      []string{"cc@example.com"},
        Bcc:     []string{"bcc@example.com"},
        ReplyTo: "replyto@example.com",
    }

    sent, err := client.Emails.Send(params)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    fmt.Println(sent.Id)
}