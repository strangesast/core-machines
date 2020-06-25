package main

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
)

func sample(deviceName string, from int) chan mMTConnectStreams {
	streams := make(chan mMTConnectStreams)
	query := make(url.Values)

	u, err := url.Parse(baseURL + "/sample")
	handleErr(err)

	if deviceName != "" {
		query.Set("path", fmt.Sprintf(`//Devices/Device[@name='%s']`, deviceName))
	}
	// `//Devices/Device[@name='%s']/Components/Linear/*/*[@name='Zact' or @name='Yact' or @name='Xact']`
	// `//Devices/Device[@name='%s']/Components/*[@name="path_basic"]/*/*[@name="execution"]'
	if from != 0 {
		query.Set("from", strconv.Itoa(from))
	}
	query.Set("interval", "0")

	u.RawQuery = query.Encode()
	resp, err := makeRequest(u)
	handleErr(err)

	decoder := xml.NewDecoder(resp.Body)

	go func() {
		for {
			t, _ := decoder.Token()

			if t == nil {
				close(streams)
				resp.Body.Close()
				break
			}

			switch se := t.(type) {
			case xml.StartElement:
				if se.Name.Local == "MTConnectStreams" {
					var s mMTConnectStreams
					decoder.DecodeElement(&s, &se)
					streams <- s
				} else if se.Name.Local == "MTConnectError" {
					var s mMTConnectError
					decoder.DecodeElement(&s, &se)
					fmt.Println("error")
					fmt.Printf("%+v\n", s)
				}
			}
		}
	}()

	return streams
}
