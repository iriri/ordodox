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
 <form action="/{{.Board}}/submit" method="POST" enctype="multipart/form-data">
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
  {{range .Threads -}}
  {{with index . 0 -}}
  <li><a href="{{.Id}}">{{.Id}}</a></li>
  {{- end}}
  {{- end}}
 </ul>
</body>`
