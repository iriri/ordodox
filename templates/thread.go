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
  <table>
   <tr><th><span>name</span></th><td><input type="text" name="name"></td></tr>
   <tr><th><span>email</span></th><td><input type="text" name="email"></td></tr>
   <tr><th><span>subject</span></th>
    <td>
     <input type="text" name="subject" style="width: 326px"><button><span>reply</span></button>
    </td>
   </tr>
   <tr><th><span>comment</span></th><td><textarea name="comment" rows="6"></textarea></td></tr>
   <tr><th><span>image</span></th>
    <td>
     <label id="upload">
      <input type="file" name="image" accept=".gif,.jpg,.jpeg,.png,image/gif,image/jpeg,image/png">
      <span>browse</span>
     </label><input type="text" name="alt" style="width: 316px">
    <td>
   </tr>
  </table>
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
    <img src="/thumb/{{.Image.Uri}}.thumb.jpg" alt="{{.Image.Alt}}" href="/img/{{.Image.Uri}}">
   </a></li>
   {{- end}}
   -----<br>
  </ul></li>
  {{- end}}
 </ul>
</body>`
