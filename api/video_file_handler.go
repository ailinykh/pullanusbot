package api

import "github.com/ailinykh/pullanusbot/v2/core"

type IVdeoFileHandler interface {
	HandleVideoFile(*core.VideoFile, core.IBot) error
}
