package templates

import (
	"html/template"
	"net/http"
)

var Index = template.Must(template.New("index").Parse(index)).Execute
var Board = template.Must(template.New("board").Parse(board)).Execute
var Thread = template.Must(template.New("thread").Parse(thread)).Execute
var Error = template.Must(template.New("error").Parse(error_)).Execute

func Reset(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Write(reset)
}

func Ordodox(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Write(ordodox)
}
