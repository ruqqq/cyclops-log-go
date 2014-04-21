package main

import (
	"fmt"
	"net/smtp"
	"strconv"
)

type Mail struct {
	Name     string
	Email    string
	Password string
	To       string
	Address  string
	Port     int
}

func (mail *Mail) send(body string) {
	auth := smtp.PlainAuth(
		"",
		mail.Email,
		mail.Password,
		mail.Address,
	)

	fmt.Println("Sending email to " + mail.To)

	err := smtp.SendMail(
		mail.Address+":"+strconv.Itoa(mail.Port),
		auth,
		mail.Email,
		[]string{mail.To},
		[]byte(body),
	)

	check(err)
}
