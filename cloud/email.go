package cloud

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mailjet/mailjet-apiv3-go"
)

// Email sends an email
func Email(w http.ResponseWriter, req *http.Request) {
	publicKey := os.Getenv("MJ_APIKEY_PUBLIC")
	secretKey := os.Getenv("MJ_APIKEY_PRIVATE")

	// topGainers := []string{stocks.Stocks}
	SenderEmail := "ochoa.erick.d@gmail.com"
	RecpientEmail := "ochoa.erick.d@gmail.com"
	emailSubject := "subject goes hur"
	emailText := "body goes hur"

	mailjetClient := mailjet.NewMailjetClient(publicKey, secretKey)

	email := &mailjet.InfoSendMail{
		FromEmail: SenderEmail,
		FromName:  "team simpl",
		Subject:   emailSubject,
		TextPart:  emailText,
		To:        RecpientEmail,
	}

	res, err := mailjetClient.SendMail(email)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.Sent[0])
	}

	data := struct {
		EmailText string
		Recipient string
	}{
		EmailText: emailText,
		Recipient: email.To,
	}

	fmt.Fprintln(w, data)
}
