package http

import "net/http"

type proxyWriter struct {
	header      func() http.Header
	write       func([]byte) (int, error)
	writeHeader func(statusCode int)
}

func (w proxyWriter) Header() http.Header {
	return w.header()
}

func (w proxyWriter) Write(b []byte) (int, error) {
	return w.write(b)
}

func (w proxyWriter) WriteHeader(statusCode int) {
	w.writeHeader(statusCode)
}