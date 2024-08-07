package test_helpers

import "fmt"

func CreateLocalizer(data map[string]string) *FakeLocalizer {
	return &FakeLocalizer{data}
}

type FakeLocalizer struct {
	data map[string]string
}

func (l *FakeLocalizer) I18n(lang, key string, args ...interface{}) string {
	if val, ok := l.data[key]; ok {
		return fmt.Sprintf(val, args...)
	}
	return key
}

func (l *FakeLocalizer) AllKeys() []string {
	keys := make([]string, 0, len(l.data))
	for k := range l.data {
		keys = append(keys, k)
	}
	return keys
}
