package twilio

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
)

func SendWhatsApp(country, phone, message string) (et.Json, error) {
	twilioSID := envar.EnvarStr("", "TWILIO_SID")
	twilioAUT := envar.EnvarStr("", "TWILIO_AUT")
	twilioFrom := envar.EnvarStr("", "TWILIO_FROM")

	apiURL := strs.Format(`https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json`, twilioSID)
	from := strs.Format(`whatsapp:%s`, twilioFrom)
	body := message
	to := strs.Format(`whatsapp:%s%s`, country, phone)

	data := url.Values{}
	data.Set("From", from)
	data.Set("Body", body)
	data.Set("To", to)

	client := &http.Client{}
	params := bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest("POST", apiURL, params)
	if err != nil {
		return et.Json{}, err
	}

	req.SetBasicAuth(twilioSID, twilioAUT)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return et.Json{}, err
	}

	defer resp.Body.Close()

	return et.Json{
		"status": resp.Status,
		"body":   resp.Body,
	}, nil
}
