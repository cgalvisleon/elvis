package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/cgalvisleon/elvis/cache"
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

func SendSMS(country string, mobile string, message string) (bool, interface{}, error) {
	var result bool

	phoneNumber := country + mobile
	sess := AwsSession()
	svc := sns.New(sess)

	message = strs.RemoveAcents(message)
	params := &sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(phoneNumber),
	}

	output, err := svc.Publish(params)
	if err != nil {
		return result, output, console.Error(err)
	}

	return true, output, nil
}

/**
* VerifyMobile
* Send sms message a code to six digit from validate user identity
**/
func VerifyMobile(app string, device string, country string, phoneNumber string) error {
	code := utility.GetCodeVerify(6)
	cache.SetVerify(device, country+phoneNumber, code)

	message := strs.Format(msg.MSG_MOBILE_VALIDATION, app, code)
	_, _, err := SendSMS(country, phoneNumber, message)
	if err != nil {
		return err
	}

	return nil
}

/**
* CheckMobile
* Check code in cache db
**/
func CheckMobile(device string, country string, mobile string, code string) (bool, error) {
	val, err := cache.GetVerify(device, country+mobile)
	if err != nil {
		return false, err
	}

	result := val == code
	if result {
		cache.DelVerify(device, country+mobile)
	}

	return result, nil
}
