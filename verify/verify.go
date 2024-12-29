package verify

import "github.com/cgalvisleon/elvis/et"

func VerifyMobille(app string, device string, country string, phoneNumber string) error {

	return nil
}

func VerifyEmail(app string, device string, email string) error {

	return nil
}

// VerifyCode
// Verify code from email or mobile
// Structure atrib is type Json this one format:
//
//	Is type email format:
//	{
//	  "kind": "email";
//	  "data": {
//	    "email": "
//	  }
//	}
//
//	Is type mobile format:
//	{
//	  "kind": "mobile";
//	  "data": {
//	    "country": "+57",
//	    "phoneNumber": "3001234567"
//	  }
//	}
func VerifyCode(app string, device string, atrib et.Json) error {
	kind := atrib.Str("kind")

	if kind == "email" {
		email := atrib.Str("email")
		return VerifyEmail(app, device, email)
	}

	if kind == "mobile" {
		country := atrib.Str("country")
		phoneNumber := atrib.Str("phoneNumber")
		return VerifyMobille(app, device, country, phoneNumber)
	}

	return nil
}

func CheckEmail(app string, device string, email, code string) (bool, error) {
	return true, nil
}

func CheckMobile(app string, device string, country, phoneNumber, code string) (bool, error) {
	return true, nil
}

// Check code
// Check code from email or mobile
// Structure atrib is type Json this one format:
//
//	Is type email format:
//	{
//	  "kind": "email";
//	  "data": {
//	    "email": "
//	  }
//	}
//
//	Is type mobile format:
//	{
//	  "kind": "mobile";
//	  "data": {
//	    "country": "+57",
//	    "phoneNumber": "3001234567"
//	  }
//	}
func CheckCode(app string, device string, atrib et.Json) (bool, error) {
	kind := atrib.Str("kind")

	if kind == "email" {
		email := atrib.Str("email")
		code := atrib.Str("code")
		return CheckEmail(app, device, email, code)
	}

	if kind == "mobile" {
		country := atrib.Str("country")
		phoneNumber := atrib.Str("phoneNumber")
		code := atrib.Str("code")
		return CheckMobile(app, device, country, phoneNumber, code)
	}

	return false, nil
}
