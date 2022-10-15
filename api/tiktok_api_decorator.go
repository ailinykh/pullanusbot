package api

func CreateTikTokAPIDecorator(primary ITikTokAPI, secondary ITikTokAPI) ITikTokAPI {
	return &TikTokAPIDecorator{primary, secondary}
}

type TikTokAPIDecorator struct {
	primary   ITikTokAPI
	secondary ITikTokAPI
}

func (api *TikTokAPIDecorator) GetItem(username string, videoId string) (*TikTokV1ItemStruct, error) {
	item, err := api.primary.GetItem(username, videoId)
	if err != nil {
		return api.secondary.GetItem(username, videoId)
	}
	return item, nil
}
