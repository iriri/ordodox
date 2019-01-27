package templates

const thread = `<!DOCTYPE html>
<html lang="en-us">
<head>
 <meta charset="utf-8">
 <meta name="viewport" content="width=device-width, initial-scale=1">
 <link type="text/css" rel="stylesheet" href="/css/ordodox.css">
 <title>/{{.Board}}/ - {{.Op.Id}}</title>
</head>
<body>
 <div class="header">
  <h1 id="title"><a href="/{{.Board}}/">/{{.Board}}/ - {{.Title}}</a></h1>
  <form action="/{{.Board}}/{{.Op.Id}}/reply" method="POST" enctype="multipart/form-data">
   <table>
    <tr><th><span class="d2">name</span></th><td><input type="text" name="name"></td></tr>
    <tr><th><span class="d2">email</span></th><td><input type="text" name="email"></td></tr>
    <tr><th><span class="d2">subject</span></th>
     <td>
      <input type="text" name="subject" style="width: 326px"><button id="submit"
      class="button" style="width: 52px"><span class="d2">reply</span></button>
     </td>
    </tr>
    <tr><th><span class="d2">comment</span></th>
     <td><textarea name="comment" rows="6"></textarea></td>
    </tr>
    <tr><th><span class="d2">image</span></th>
     <td>
      <label id="upload" class="button" style="width: 62px">
       <input type="file" name="image" accept=".gif,.jpg,.jpeg,.png,image/gif,image/jpeg,image/png">
       <span class="d2">browse</span>
      </label><input type="text" name="alt" style="width: 314px">
     <td>
    </tr>
   </table>
  </form>
 </div>
 <div>
  <div id="{{.Op.Id}}" class="post op">
   <div>
    {{if .Op.Subject -}}
    <span class="subject">{{.Op.Subject}}</span>
    {{- end}}
    <span class="name">
     {{- if .Op.Email -}}
     <a href="mailto:{{.Op.Email}}">{{.Op.Name}}</a>
     {{- else -}}
     {{.Op.Name}}
     {{- end -}}
    </span>
    <span class="date">{{.Op.Date}}</span>
    <span class="id">No. {{.Op.Id}}</span>
   {{if .Op.Image -}}
    <div class="imageinfo">
     <a href="/img/{{.Op.Image.Uri}}">{{.Op.Image.Name}}</a>
     {{.Op.Image.Width}}x{{.Op.Image.Height}}
     {{.Op.Image.Size}}k
    </div>
   </div>
   <div class="image">
    <a href="/img/{{.Op.Image.Uri}}">
     <img src="/thumb/{{.Op.Image.Uri}}.thumb.jpg" alt="{{.Op.Image.Alt}}" href="/img/{{.Op.Image.Uri}}">
    </a>
   </div>
   {{- else -}}
   </div>
   {{- end}}
   <div class="comment">{{.Op.Comment}}</div>
  </div>
 </div>
 {{range .Posts -}}
 <div>
  <div class="post reply">
   <div>
    {{if .Subject -}}
    <span class="subject">{{.Subject}}</span>
    {{- end}}
    <span class="name">
     {{- if .Email -}}
     <a href="mailto:{{.Email}}">{{.Name}}</a>
     {{- else -}}
     {{.Name}}
     {{- end -}}
    </span>
    <span class="date">{{.Date}}</span>
    <span class="id">No. {{.Id}}</span>
   {{if .Image -}}
    <div class="imageinfo">
     <a href="/img/{{.Image.Uri}}">{{.Image.Name}}</a>
     {{.Image.Width}}x{{.Image.Height}}
     {{.Image.Size}}k
    </div>
   </div>
   <div class="image">
    <a href="/img/{{.Image.Uri}}">
     <img src="/thumb/{{.Image.Uri}}.thumb.jpg" alt="{{.Image.Alt}}" href="/img/{{.Image.Uri}}">
    </a>
   </div>
   {{- else -}}
   </div>
   {{- end}}
   <div class="comment">{{.Comment}}</div>
  </div>
 </div>
 {{- end}}
</body>`
