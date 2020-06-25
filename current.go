package main

import (
	"encoding/xml"
	"fmt"
	"net/url"
)

func current(deviceName string) mMTConnectStreams {
	u, err := url.Parse(baseURL + "/current")
	query := make(url.Values)

	if deviceName != "" {
		query.Set("path", fmt.Sprintf(`//Devices/Device[@name='%s']`, deviceName))
	}
	u.RawQuery = query.Encode()

	handleErr(err)

	resp, err := makeRequest(u)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)

	var s mMTConnectStreams
	for {
		t, _ := decoder.Token()

		if t == nil {
			resp.Body.Close()
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "MTConnectStreams" {

				decoder.DecodeElement(&s, &se)
			} else if se.Name.Local == "MTConnectError" {
				var s mMTConnectError
				decoder.DecodeElement(&s, &se)
				fmt.Println("error")
				fmt.Printf("%+v\n", s)
			}
		}
	}
	return s
}
