package api

type ITikTokAPI interface {
	GetItem(string, string) (*TikTokItem, error)
}

type TikTokItem struct {
	Author   TikTokAuthor
	Desc     string
	Music    TikTokMusic
	Stickers []string
	VideoURL string
}

type TikTokAuthor struct {
	Nickname string
	UniqueId string
}

type TikTokMusic struct {
	Title string
}
