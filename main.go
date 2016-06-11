package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func doRequest(requestType string, client *http.Client, url *url.URL) (response *http.Response, err error) {
	var request *http.Request
	request, err = http.NewRequest(requestType, url.String(), nil)
	if err == nil {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		response, err = client.Do(request)
	}

	return
}

func debug(res *http.Response) {
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Printf("Statuscode: %v\n", res.StatusCode)
	fmt.Println()
	for k, v := range res.Header {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Println()
	fmt.Printf("%s\n", body)
}

func main() {
	p := NewPremiumize()
	res := p.GetDownloadLinks()
	for _, d := range res {
		fmt.Println(d)
	}
}
