package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"

	"golang.org/x/net/html"
)

type Download struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Type      string `json:"type"`
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	Size      string `json:"size"`
}

type Premiumize struct {
	DownloadList []Download `json:"content"`
	Client       *http.Client
	Urls         map[string]*url.URL
	Endpoints    map[string]string
	LoginCookie  string `json:"loginCookie"`
}

func loadConfigurationFile(filename string) (file []byte, err error) {
	file, err = ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Cannot read from config file: %v\n", err)
		os.Exit(1)
	}
	return
}

func (premiumize *Premiumize) initializeClient(urls map[string]*url.URL) {
	var cookies []*http.Cookie
	premiumize.Client = &http.Client{}
	cookies = append(cookies, &http.Cookie{Name: "login", Value: premiumize.LoginCookie})
	premiumize.Client.Jar, _ = cookiejar.New(nil)
	for _, url := range urls {
		premiumize.Client.Jar.SetCookies(url, cookies)
	}
}

func NewPremiumize() (premiumize *Premiumize) {
	data, _ := loadConfigurationFile("config.json")
	json.Unmarshal(data, &premiumize)
	folderListUrl, _ := url.Parse("https://www.premiumize.me/api/folder/list")
	detailUrl, _ := url.Parse("https://www.premiumize.me/browsetorrent")
	premiumize.Urls = map[string]*url.URL{
		"folderListUrl": folderListUrl,
		"detailUrl":     detailUrl,
	}
	premiumize.initializeClient(premiumize.Urls)
	return
}

func parseDetailPage(responseBody io.ReadCloser) (hrefs []string) {
	z := html.NewTokenizer(responseBody)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if isAnchor {
				for _, a := range t.Attr {
					if a.Key == "href" {
						re := regexp.MustCompile("^https.*(mkv|mp4)")
						if re.FindStringSubmatch(a.Val) != nil {
							hrefs = append(hrefs, a.Val)
						}
					}
					break
				}
			}
		}
	}
}

func (premiumize *Premiumize) getDownloadList() {
	res, _ := doRequest("POST", premiumize.Client, premiumize.Urls["folderListUrl"])
	responseBody, err := ioutil.ReadAll(res.Body)
	if err == nil {
		defer res.Body.Close()
	}
	json.Unmarshal(responseBody, &premiumize)

}

func (premiumize *Premiumize) GetDownloadLinks() (downloadLinks []string) {
	premiumize.getDownloadList()
	for _, download := range premiumize.DownloadList {
		detailUrl := premiumize.Urls["detailUrl"]
		q := detailUrl.Query()
		q.Set("hash", download.Hash)
		detailUrl.RawQuery = q.Encode()
		res, _ := doRequest("GET", premiumize.Client, detailUrl)
		hrefs := parseDetailPage(res.Body)
		downloadLinks = append(downloadLinks, hrefs...)
	}
	return
}
