package helpers

import (
	"os"
	"path"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateDownloadVideoFactory(l core.ILogger, fileDownloader core.IFileDownloader, videoFactory core.IVideoFactory) core.IVideoFactory {
	return &DownloadVideoFactory{l, fileDownloader, videoFactory}
}

type DownloadVideoFactory struct {
	l              core.ILogger
	fileDownloader core.IFileDownloader
	videoFactory   core.IVideoFactory
}

// CreateVideo is a core.IVideoFactory interface implementation
func (factory *DownloadVideoFactory) CreateVideo(url string) (*core.Video, error) {
	filename := path.Base(url)
	if strings.Contains(filename, "?") {
		parts := strings.Split(url, "?")
		filename = path.Base(parts[0])
	}

	if !strings.HasSuffix(filename, ".mp4") {
		filename = filename + ".mp4"
	}

	videoPath := path.Join(os.TempDir(), filename)
	file, err := factory.fileDownloader.Download(url, videoPath)
	if err != nil {
		factory.l.Error(err)
		file.Dispose()
		return nil, err
	}
	factory.l.Infof("file downloaded: %s %0.2fMB", file.Name, float64(file.Size)/1024/1024)
	return factory.videoFactory.CreateVideo(file.Path)
}
