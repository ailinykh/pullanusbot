package core

type SettingKey string

const (
	SFaggotGameEnabled         SettingKey = "faggot_game"
	SInstagramFlowEnabled      SettingKey = "instagram_flow"
	SInstagramFlowRemoveSource SettingKey = "instagram_flow_remove_source"
	SLinkFlowEnabled           SettingKey = "link_flow"
	SLinkFlowRemoveSource      SettingKey = "link_flow_remove_source"
	SPayloadList               SettingKey = "payload_list"
	STwitterFlowEnabled        SettingKey = "twitter_flo"
	STwitterFlowRemoveSource   SettingKey = "twitter_flow_remove_source"
	SYoutubeFlowEnabled        SettingKey = "youtube_flow"
	SYoutubeFlowRemoveSource   SettingKey = "youtube_flow_remove_source"
)
