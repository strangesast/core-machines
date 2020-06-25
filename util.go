package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func handleErr(err interface{}) {
	if err != nil {
		panic(err)
	}
}

func pprint(arg interface{}) {
	b, _ := json.MarshalIndent(arg, "", "  ")
	fmt.Println(string(b))
}

type mHTTPError struct {
	URL *url.URL
	Err error
}

func (e *mHTTPError) Error() string {
	return "http request failed for " + e.URL.String() + ": " + e.Err.Error()
}

func makeRequest(u *url.URL) (*http.Response, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(`bad response`)
		err = &mHTTPError{u, nil}
		return nil, err
	}
	return resp, nil
}
