package api

type TikTokV2HTMLNResponse struct {
	ItemModule map[string]*TikTokV2Item
	UserModule TikTokV2UserModule
	VideoPage  TikTokV2VideoPage
}

type TikTokV2Item struct {
	Desc           string
	Author         string
	Music          TikTokV1Music
	Video          TikTokV1Video
	StickersOnItem []TikTokV1Sticker
}

type TikTokV2UserModule struct {
	Users map[string]*TikTokAuthor
}

type TikTokV2VideoPage struct {
	StatusCode int
}
