package app

import (
	"errors"

	"github.com/badoux/checkmail"
)

type MData struct {
	Email string
}

func (data *MData) Validate() (string, error) {
	email := data.Email
	err := checkmail.ValidateFormat(email)
	if err != nil {
		return "", errors.New("Invalid email format: " + err.Error())
	}
	err = checkmail.ValidateHost(email)
	if err != nil {
		return "", errors.New("Invalid email host: " + err.Error())
	}
	domain := "ybenel.cf"
	sender := "root@ybenel.cf"
	err = checkmail.ValidateHostAndUser(domain, sender, email)
	if err != nil {
		return "", errors.New("Invalid email: " + err.Error())
	}
	return "Valid email", nil
}
