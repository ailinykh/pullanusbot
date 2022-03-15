package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateCookieJar(l core.ILogger, cookiesFile string) *CookieJar {
	return &CookieJar{l, cookiesFile, make(map[string][]*http.Cookie)}
}

type CookieJar struct {
	l        core.ILogger
	filename string
	cookies  map[string][]*http.Cookie
}

func (jar *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies[u.Host] = jar.merge(jar.cookies[u.Host], cookies)
	data, err := json.MarshalIndent(jar.cookies, "", "  ")
	if err != nil {
		jar.l.Error(err)
		return
	}

	err = ioutil.WriteFile(jar.filename, data, 0644)
	if err != nil {
		jar.l.Error(err)
		return
	}
}

func (jar *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	data, err := ioutil.ReadFile(jar.filename)
	if err != nil {
		jar.l.Error(err)
		jar.cookies = map[string][]*http.Cookie{}
		return []*http.Cookie{}
	}

	err = json.Unmarshal(data, &jar.cookies)
	if err != nil {
		jar.l.Error(err)
		jar.cookies = map[string][]*http.Cookie{}
		return []*http.Cookie{}
	}

	return jar.cookies[u.Host]
}

func (j *CookieJar) merge(lhs []*http.Cookie, rhs []*http.Cookie) []*http.Cookie {
	for _, r := range rhs {
		found := false
		for _, l := range lhs {
			if l.Name == r.Name {
				found = true
				l = r
				break
			}
		}
		if !found {
			lhs = append(lhs, r)
		}
	}
	return lhs
}
