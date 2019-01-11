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
 <form action="/{{.Board}}/{{.Op}}/reply" method="POST" enctype="multipart/form-data">
  name: <input type="text" name="name"><br>
  email: <input type="text" name="email"><br>
  subject: <input type="text" name="subject"><br>
  comment: <input type="text" name="comment"><br>
  image: <input type="file" name="image" accept=".gif,.jpg,.jpeg,.png,image/gif,image/jpeg,image/png"><br>
  <input type="submit" value="reply">
 </form>
 <ul>
  {{range .Posts -}}
  <li><ul>
   <li>id: {{.Id}}</li>
   <li>date: {{.Date}}</li>
   <li>name: {{.Name}}</li>
   <li>email: {{.Email}}</li>
   <li>subject: {{.Subject}}</li>
   <li>comment: {{.Comment}}</li>
   {{if .Image -}}
   <li>filename: {{.Image.Name}}</li>
   <li><a href="/img/{{.Image.Uri}}">
    <img src="/thumb/{{.Image.Uri}}.thumb.jpg" href="/img/{{.Image.Uri}}">
   </a></li>
   {{- end}}
   -----<br>
  </ul></li>
  {{- end}}
 </ul>
</body>`
