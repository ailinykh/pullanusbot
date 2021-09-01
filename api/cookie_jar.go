package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type CookieJar struct {
	cookies map[string][]*http.Cookie
}

func (j *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.cookies[u.Host] = cookies
	data, err := json.MarshalIndent(j.cookies, "", "  ")
	if err != nil {
		fmt.Println("❌⚠️", err)
		return
	}
	filename := "cookies.json"
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("❌⚠️", err)
		return
	}
}

func (j *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	filename := "cookies.json"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("⚠️", err)
		j.cookies = map[string][]*http.Cookie{}
		return []*http.Cookie{}
	}

	err = json.Unmarshal(data, &j.cookies)
	if err != nil {
		fmt.Println("⚠️", err)
		j.cookies = map[string][]*http.Cookie{}
		return []*http.Cookie{}
	}

	return j.cookies[u.Host]
}
