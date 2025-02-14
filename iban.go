// Package iban contains utility functions for working with IBAN account numbers.
package iban

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidIBAN is returned when an invalid IBAN number was received
var ErrInvalidIBAN = errors.New("Invalid IBAN number received")

// IBAN represents an IBAN number, split up into its different parts.
type IBAN struct {
	Number         string
	NumberFormated string
	CountryCode    string
	Checksum       string
	BBAN           string
	BankCode       string
	SortCode       string
	AccountNumber  string
}

// MarshalText implements encoding/text TextMarshaler interface.
func (i IBAN) MarshalText() ([]byte, error) {
	return []byte(i.Number), nil
}

// UnmarshalText implements encoding/text TextUnmarshaler interface.
func (i *IBAN) UnmarshalText(text []byte) error {
	iban, err := NewIBAN(string(text))
	if err != nil {
		return err
	}
	*i = iban
	return nil
}

// NewIBAN creates a new instance of IBAN and checks before if the IBAN
// number is actually valid or not.
func NewIBAN(ibanNumber string) (IBAN, error) {
	var newIBAN IBAN

	isIBANValid, formatedIBANNumber, err := IsCorrectIban(ibanNumber, false)
	if !isIBANValid {
		return IBAN{}, ErrInvalidIBAN
	}

	if err != nil {
		return IBAN{}, err
	}

	newIBAN.Number = removeSpaces(ibanNumber)
	newIBAN.NumberFormated = formatedIBANNumber
	newIBAN.CountryCode, newIBAN.Checksum, newIBAN.BBAN = splitIbanUp(formatedIBANNumber)
	newIBAN.BankCode, newIBAN.SortCode, newIBAN.AccountNumber = getBankInfo(newIBAN)

	return newIBAN, nil
}

func getBankInfo(iban IBAN) (bankCode string, sortCode string, accountNumber string) {
	bankCode = ""
	sortCode = ""
	accountNumber = ""

	for _, code := range countryList {
		if code.code == iban.CountryCode {
			// Bank code
			bFirstIndex := strings.Index(code.ibanFields, "b")
			bLastIndex := strings.LastIndex(code.ibanFields, "b")
			if bFirstIndex >= 0 && bLastIndex >= 0 {
				bankCode = removeSpaces(iban.NumberFormated[bFirstIndex : bLastIndex+1])
			}

			// Sort code
			sFirstIndex := strings.Index(code.ibanFields, "s")
			sLastIndex := strings.LastIndex(code.ibanFields, "s")
			if sFirstIndex >= 0 && sLastIndex >= 0 {
				sortCode = removeSpaces(iban.NumberFormated[sFirstIndex : sLastIndex+1])
			}

			// Account number
			cFirstIndex := strings.Index(code.ibanFields, "c")
			cLastIndex := strings.LastIndex(code.ibanFields, "c")
			if cFirstIndex >= 0 && cLastIndex >= 0 {
				accountNumber = removeSpaces(iban.NumberFormated[cFirstIndex : cLastIndex+1])
			}

			return bankCode, sortCode, accountNumber
		}
	}

	return bankCode, sortCode, accountNumber
}

type ibanCountry struct {
	country           string
	chars             int
	bbanFormat        string
	code              string
	ibanFields        string
	comment           string
	standardTreatment bool
}

// IsCorrectIban checks if the given iban number corresponds to the rules of a valid iban number and also
// returns a properly formated iban number string.
func IsCorrectIban(iban string, debug bool) (isValid bool, wellFormated string, err error) {
	var ibanConfig ibanCountry
	ibanValid := false
	wellFormated = ""
	if len(iban) >= 15 { // Minimum length for IBAN

		// Clean up string
		iban = strings.ToUpper(strings.Replace(iban, " ", "", -1))

		// Split string up
		passedChars := len(iban)
		passedCode, passedChecksum, passedBban := splitIbanUp(iban)

		ibanConfig = countryList[passedCode]

		if ibanConfig.chars > 0 {
			if ibanConfig.chars == passedChars {
				convertedIban := rearrangeIBAN(passedCode, passedChecksum, passedBban)
				convertedIban = convertCharToNumber(convertedIban)

				if calculateModulo(convertedIban) == 1 {
					ibanValid = true
					wellFormated = splitTo4(iban)
				}
			} else {
				return false, wellFormated, fmt.Errorf("IBAN: lenght (%s) does not match configuration length (%s)", strconv.Itoa(passedChars), strconv.Itoa(ibanConfig.chars))
			}

		} else {
			return false, wellFormated, fmt.Errorf("IBAN: country <%s> is not in the list", passedCode)
		}

	} else {
		return false, wellFormated, fmt.Errorf("IBAN: incorect IBAN string passed <%s>", iban)
	}
	return ibanValid, wellFormated, nil
}

func splitIbanUp(iban string) (countryCode string, checksum string, bban string) {
	countryCode = iban[0:2]
	checksum = iban[2:4]
	bban = iban[4:]
	return countryCode, checksum, bban
}

func splitTo4(value string) (returnValue string) {
	n := 0
	for n < len(value) {

		if n < len(value)-3 {
			returnValue += value[n:n+4] + " "
		} else {
			returnValue += value[n:]
		}
		n += 4
	}
	return strings.TrimSpace(returnValue)
}

// GetIbanChecksum returns the checksum of the given iban number.
func GetIbanChecksum(iban string) (int, error) {
	ibanChecksum := 0

	if len(iban) > 15 { // Minimum length for IBAN

		iban = strings.ToUpper(strings.Replace(iban, " ", "", -1))

		// Split string up
		passedCode, passedChecksum, passedBban := splitIbanUp(iban)
		passedChecksum = "00"
		convertedIban := rearrangeIBAN(passedCode, passedChecksum, passedBban)
		convertedIban = convertCharToNumber(convertedIban)

		ibanChecksum = 98 - calculateModulo(convertedIban)

	} else {
		return -1, fmt.Errorf("IBAN: incorect IBAN string passed <%s>", iban)
	}
	return ibanChecksum, nil
}

func convertCharToNumber(value string) string {
	var temp = []byte(value)
	var result string

	for _, item := range temp {
		if !((item < "A"[0]) || (item > "Z"[0])) {
			item = byte(item - 55)
			result += strconv.Itoa(int(item))
		} else {
			result += string(item)
		}
	}
	return result
}

func rearrangeIBAN(countryCode string, checksum string, bban string) string {
	return bban + countryCode + checksum
}

func calculateModulo(iban string) int {
	exit := false
	tempIBAN := ""
	rest := 0
	for !exit {
		if len(iban) > 9 {
			tempIBAN = iban[0:9]

			value, err := strconv.Atoi(tempIBAN)
			if err != nil {
				exit = true
				return -1
			}

			rest = (value % 97)
			iban = strconv.Itoa(rest) + iban[9:]
		} else {
			tempIBAN = iban

			value, err := strconv.Atoi(tempIBAN)
			if err != nil {
				exit = true
				return -1
			}

			rest = (value % 97)
			exit = true
		}
	}
	return rest
}

func removeSpaces(text string) string {
	return strings.ReplaceAll(text, " ", "")
}
