package http

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// CacheHandler returns a decorated version of the given cache that injects
// cache related headers.
func CacheHandler(h http.Handler, webDir string, maxAge time.Duration) http.Handler {
	return &cacheHandler{
		Handler: h,
		maxAge:  maxAge,
		webDir:  webDir,
	}
}

type cacheHandler struct {
	http.Handler

	once         sync.Once
	etag         string
	cacheControl string
	maxAge       time.Duration
	webDir       string
}

func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.once.Do(h.init)

	if r.URL.Path == "/.etag" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.etag == "" {
		h.Handler.ServeHTTP(w, r)
		return
	}

	w.Header().Set("ETag", h.etag)
	w.Header().Set("Cache-Control", h.cacheControl)

	etag := r.Header.Get("If-None-Match")
	if etag == h.etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	h.Handler.ServeHTTP(w, r)
}

func (h *cacheHandler) init() {
	h.etag = h.getEtag()
	h.cacheControl = h.getCacheControl()
}

func (h *cacheHandler) getEtag() string {
	filename := filepath.Join(h.webDir, ".etag")

	etag, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(etag)
}

func (h *cacheHandler) getCacheControl() string {
	if h.maxAge > 0 {
		return fmt.Sprintf("private, max-age=%.f", h.maxAge.Seconds())
	}
	return "no-cache"
}

// GenerateEtag generates an etag.
func GenerateEtag() string {
	t := time.Now().UTC().String()
	return fmt.Sprintf(`"%x"`, sha1.Sum([]byte(t)))
}
