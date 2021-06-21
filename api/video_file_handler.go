package api

import "github.com/ailinykh/pullanusbot/v2/core"

// IVdeoFileHandler interface for processing Videos
type IVdeoFileHandler interface {
	HandleVideo(*core.Video, core.IBot) error
}
