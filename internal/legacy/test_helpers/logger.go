package test_helpers

func CreateLogger() *FakeLogger {
	return &FakeLogger{}
}

type FakeLogger struct{}

func (FakeLogger) Close()                          {}
func (FakeLogger) Error(...interface{})            {}
func (FakeLogger) Errorf(string, ...interface{})   {}
func (FakeLogger) Info(...interface{})             {}
func (FakeLogger) Infof(string, ...interface{})    {}
func (FakeLogger) Warning(...interface{})          {}
func (FakeLogger) Warningf(string, ...interface{}) {}
