package templates

import "html/template"

var Index = template.Must(template.New("index").Parse(index)).Execute
var Board = template.Must(template.New("board").Parse(board)).Execute
var Thread = template.Must(template.New("thread").Parse(thread)).Execute
var Error = template.Must(template.New("error").Parse(error_)).Execute
