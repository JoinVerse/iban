package iban_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/JoinVerse/iban"
	"github.com/stretchr/testify/assert"
)

var validIBANTestNumbers = []struct {
	number   string
	bankCode string
	sortCode string
}{
	{"LU28 0019 4006 4475 0000", "001", ""},
	{"ES9121000418450200051332", "2100", ""},
	{"ES3502297205860300042630", "0229", ""},
	{"ES9121000418450200051332       ", "2100", ""},
}

var invalidIBANTestNumbers = []struct {
	number string
}{
	{"LU12 3456 7890 1234 5678"},
}

func TestValidIBAN(t *testing.T) {
	for _, ibanTestNumber := range validIBANTestNumbers {
		result, err := iban.NewIBAN(ibanTestNumber.number)
		if err != nil || result == (iban.IBAN{}) {
			t.Error("No object was created!")
			t.Log(err)
		}

		assert.Equal(t, ibanTestNumber.bankCode, result.BankCode)
		assert.Equal(t, ibanTestNumber.sortCode, result.SortCode)
	}
}

func TestInvalidIBAN(t *testing.T) {
	for _, ibanTestNumber := range invalidIBANTestNumbers {
		_, err := iban.NewIBAN(ibanTestNumber.number)
		if err == nil {
			t.Error("No error was thrown for an invalid IBAN number!")
		}
	}
}

type IBANMessage struct {
	Country string `json:"country"`
	Code    string `json:"code"`
	IBAN    string `json:"iban"`
}

type IBANList struct {
	IBANs []IBANMessage `json:"ibans"`
}

func TestIsCorrectIban(t *testing.T) {

	data, err := ioutil.ReadFile("./data/iban.json")
	if err != nil {
		t.Error("error reading file", err)
		t.FailNow()
	}
	var iList IBANList
	err = json.Unmarshal(data, &iList)
	if err != nil {
		t.Error("error unmarshall data into the IBAN list", err)
		t.FailNow()
	}

	for k, message := range iList.IBANs {
		ok, _, _ := iban.IsCorrectIban(message.IBAN, false)
		if !ok {
			t.Error("for test waiting for true got false", k, ":", message)
		}
	}
}
