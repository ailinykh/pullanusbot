package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/usecases"
)

func CreateTikTokAPI(l core.ILogger, hc core.IHttpClient) usecases.ITikTokAPI {
	return &TikTokAPI{l, hc}
}

type TikTokAPI struct {
	l  core.ILogger
	hc core.IHttpClient
}

func (api *TikTokAPI) Get(url string) (*usecases.TikTokHTMLResponse, error) {

	getRand := func(count int) string {
		rv := ""
		for i := 1; i < count; i++ {
			rv = rv + strconv.Itoa(rand.Intn(10))
		}
		return rv
	}
	api.hc.SetHeader("Cookie", "tt_webid_v2=69"+getRand(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	htmlString, err := api.hc.GetContent(url)
	if err != nil {
		return nil, err
	}

	// os.WriteFile("tiktok-"+strings.Split(url, "/")[5]+".html", []byte(htmlString), 0644)
	r := regexp.MustCompile(`<script id="__NEXT_DATA__" type="application\/json" nonce="[\w-]+" crossorigin="anonymous">(.*?)<\/script>`)
	match := r.FindStringSubmatch(htmlString)
	if len(match) < 1 {
		api.l.Error(match)
		return nil, fmt.Errorf("unexpected html")
	}

	var resp *usecases.TikTokHTMLResponse
	err = json.Unmarshal([]byte(match[1]), &resp)
	if err != nil {
		return nil, err
	}

	return resp, err
}
