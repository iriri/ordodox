package templates

const thread = `<!DOCTYPE html>
<html lang="en-us">
<head>
 <meta charset="utf-8">
 <meta name="viewport" content="width=device-width, initial-scale=1">
 <link type="text/css" rel="stylesheet" href="/css/reset.css">
 <link type="text/css" rel="stylesheet" href="/css/ordodox.css">
 <title>{{.Op}}</title>
</head>
<body>
 <h3>{{.Op}}</h3>
 <ul>
  {{range .Posts -}}
  <li>{{.Body}}</li>
  {{- end}}
 </ul>
</body>`
