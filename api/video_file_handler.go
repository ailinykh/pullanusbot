package api

import "github.com/ailinykh/pullanusbot/v2/core"

// IVdeoFileHandler interface for processing VideoFiles
type IVdeoFileHandler interface {
	HandleVideoFile(*core.VideoFile, core.IBot) error
}
