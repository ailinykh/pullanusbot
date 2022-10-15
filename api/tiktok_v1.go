package api

type TikTokV1JSONResponse struct {
	ItemInfo TikTokV1ItemInfo
}

type TikTokV1HTMLResponse struct {
	Props TikTokV1HTMLProps
}

type TikTokV1HTMLProps struct {
	PageProps TikTokV1Response
}

type TikTokV1Response struct {
	ServerCode int
	StatusCode int
	ItemInfo   TikTokV1ItemInfo
}

type TikTokV1ItemInfo struct {
	ItemStruct TikTokV1ItemStruct
}

type TikTokV1ItemStruct struct {
	Desc           string
	Author         TikTokV1Author
	Music          TikTokV1Music
	Video          TikTokV1Video
	StickersOnItem []TikTokV1Sticker
}

type TikTokV1Author struct {
	UniqueId string
	Nickname string
}

type TikTokV1Music struct {
	Id         string
	Title      string
	AuthorName string
}

type TikTokV1Video struct {
	Id           string
	DownloadAddr string
	ShareCover   []string
	Bitrate      int
	CodecType    string
}

type TikTokV1Sticker struct {
	StickerText []string
	StickerType int
}
