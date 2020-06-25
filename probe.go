package main

import (
	"encoding/xml"
	"io/ioutil"
	"net/url"
)

func probe() mMTConnectDevices {
	u, err := url.Parse(baseURL + "/probe")

	handleErr(err)

	resp, err := makeRequest(u)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var devices mMTConnectDevices

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	handleErr(err)

	err = xml.Unmarshal(body, &devices)
	handleErr(err)

	return devices
}
