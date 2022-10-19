package api

type IgReel struct {
	Items []IgReelItem
}

type IgUser struct {
	Username string
	FullName string `json:"full_name"`
}

type IgReelItem struct {
	Code          string
	User          IgUser
	Caption       IgCaption
	VideoDuration float64             `json:"video_duration"`
	VideoVersions []IgReelVideo       `json:"video_versions"`
	ClipsMetadata IgReelClipsMetadata `json:"clips_metadata"`
}

type IgReelVideo struct {
	Width  int
	Height int
	URL    string
}

type IgCaption struct {
	Text string
}

type IgReelClipsMetadata struct {
	MusicInfo         *IgReelMusicInfo         `json:"music_info"`
	OriginalSoundInfo *IgReelOriginalSoundInfo `json:"original_sound_info"`
}

type IgReelMusicInfo struct {
	MusicAssetInfo IgReelMusicAssetInfo `json:"music_asset_info"`
}

type IgReelMusicAssetInfo struct {
	DisplayArtist          string `json:"display_artist"`
	Title                  string
	ProgressiveDownloadURL string `json:"progressive_download_url"`
}

type IgReelOriginalSoundInfo struct {
	ProgressiveDownloadURL string `json:"progressive_download_url"`
}
