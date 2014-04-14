package main

import (
	"fmt"
	"net/smtp"
)

type Mail struct {
	Name     string
	Email    string
	Password string
	To       string
}

func (mail *Mail) send(body string) {
	auth := smtp.PlainAuth(
		"",
		mail.Email,
		mail.Password,
		"smtp.gmail.com",
	)

	fmt.Println("Sending email to " + mail.To)

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		mail.Email,
		[]string{mail.To},
		[]byte(body),
	)

	check(err)
}
