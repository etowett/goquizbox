package util

import (
	"fmt"
	"regexp"
	"strings"
)

func GetValidPhone(phoneNumber string) (string, error) {
	badChars := []string{"\t", "\n", " ", ",", "-", "(", ")", ".", "'", "\""}
	if phoneNumber == "" || len(phoneNumber) < 5 {
		return "", fmt.Errorf("phoneNumber given (%v) too short", phoneNumber)
	}
	for i := range badChars {
		phoneNumber = strings.Replace(phoneNumber, badChars[i], "", -1)
	}
	var err error
	var validNumber string
	if phoneNumber[0:1] == "+" {
		validNumber, err = isInternational(phoneNumber)
		if err != nil {
			return "", err
		}
	} else {
		validNumber, err = isKenyan(phoneNumber)
		if err != nil {
			return "", err
		}
	}
	return validNumber, nil
}

func isInternational(phoneNumber string) (string, error) {
	if phoneNumber[1:4] == "254" {
		return isKenyan(phoneNumber)
	}
	match, err := regexp.MatchString("^\\+{1}[0-9]{7,15}$", phoneNumber)
	if err != nil {
		return "", fmt.Errorf("international - error matching regex: %v", err)
	}
	if match == false {
		return "", fmt.Errorf("international - given number could not match regex")
	}
	return phoneNumber, nil
}

func isKenyan(phoneNumber string) (string, error) {
	pattern := "^[7]{1}[0-9]{8}$"

	if phoneNumber[0:1] == "+" || phoneNumber[0:1] == "0" {
		phoneNumber = phoneNumber[1:]
	}
	if phoneNumber[0:3] == "254" {
		phoneNumber = phoneNumber[3:]
	}
	if phoneNumber[0:1] == "0" {
		phoneNumber = phoneNumber[1:]
	}
	match, err := regexp.MatchString(pattern, phoneNumber)
	if err != nil {
		return "", fmt.Errorf("kenyan - error matching regex: %v", err)
	}
	if match == false {
		return "", fmt.Errorf("kenyan - given number could not match regex")
	}
	return "+254" + phoneNumber, nil
}
