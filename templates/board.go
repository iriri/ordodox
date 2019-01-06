package templates

const board = `<!DOCTYPE html>
<html lang="en-us">
<head>
 <meta charset="utf-8">
 <meta name="viewport" content="width=device-width, initial-scale=1">
 <link type="text/css" rel="stylesheet" href="/css/reset.css">
 <link type="text/css" rel="stylesheet" href="/css/ordodox.css">
 <title>/{{.Board}}/</title>
</head>
<body>
 <h3>/{{.Board}}/</h3>
 <ul>
  {{range .Threads -}}
  {{with index . 0 -}}
  <li><a href="{{.Id}}">{{.Id}}</a></li>
  {{- end}}
  {{- end}}
 </ul>
</body>`
