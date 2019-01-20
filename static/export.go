package static

import "net/http"

func Ordodox(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Write(ordodox)
}
