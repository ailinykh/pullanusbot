package xui

import "net/http"

type AddHeaderTransport struct {
	T http.RoundTripper
	H map[string]string
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range adt.H {
		req.Header.Add(k, v)
	}
	return adt.T.RoundTrip(req)
}

func NewAddHeaderTransport(T http.RoundTripper, H map[string]string) *AddHeaderTransport {
	if T == nil {
		T = http.DefaultTransport
	}
	if H == nil {
		H = make(map[string]string)
	}
	return &AddHeaderTransport{T, H}
}
