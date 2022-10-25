package core

// ILogger for logging
type ILogger interface {
	Close()
	Error(...interface{})
	Errorf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warning(...interface{})
	Warningf(string, ...interface{})
}
