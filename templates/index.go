package templates

const index = `<!DOCTYPE html>
<html lang="en-us">
<head>
 <meta charset="utf-8">
 <meta name="viewport" content="width=device-width, initial-scale=1">
 <link type="text/css" rel="stylesheet" href="/css/reset.css">
 <link type="text/css" rel="stylesheet" href="/css/ordodox.css">
 <title>boards</title>
</head>
<body>
 <h3>boards</h3>
 <ul>
  {{range . -}}
  <li><a href="{{.Name}}/">/{{.Name}}/ - {{.Title}}</a></li>
  {{- end}}
 </ul>
</body>`
