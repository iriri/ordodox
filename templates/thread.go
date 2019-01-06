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
 <form action="/{{.Board}}/{{.Op}}/reply" method="POST">
  name: <input type="text" name="name"><br>
  email: <input type="text" name="email"><br>
  subject: <input type="text" name="subject"><br>
  comment: <input type="text" name="comment"><br>
  <input type="submit" value="reply">
 </form>
 <ul>
  {{range .Posts -}}
  <li>{{.Comment}}</li>
  {{- end}}
 </ul>
</body>`
