package test_helpers

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateMediaFactory() *FakeMediaFactory {
	return &FakeMediaFactory{[]core.URL{}}
}

type FakeMediaFactory struct {
	URLs []core.URL
}

func (factory *FakeMediaFactory) CreateMedia(url core.URL) ([]*core.Media, error) {
	factory.URLs = append(factory.URLs, url)
	return []*core.Media{{URL: url}}, nil
}
