package templates

const board = `<!DOCTYPE html>
<html lang="en-us">
<head>
 <meta charset="utf-8">
 <meta name="viewport" content="width=device-width, initial-scale=1">
 <link type="text/css" rel="stylesheet" href="/css/ordodox.css">
 <title>/{{.Board}}/</title>
</head>
<body>
 <header>
  <h1 id="title">/{{.Board}}/ - {{.Title}}</h1>
  <form action="/{{.Board}}/submit" method="POST" enctype="multipart/form-data">
   <table>
    <tr><th>name</th><td><input type="text" name="name"></td></tr>
    <tr><th>email</th><td><input type="text" name="email"></td></tr>
    <tr><th>subject</th>
     <td>
      <input type="text" name="subject" style="width: 316px"><button id="submit"
      class="button" style="width: 62px"><span class="rel" style="right: 1px">submit</span></button>
     </td>
    </tr>
    <tr><th>comment</th>
     <td><textarea name="comment" rows="6" placeholder="formatting&#10; >greentext, >>reference&#10; ` +
	"`code`, ```code block```" + `&#10; \escaped"></textarea></td>
    </tr>
    <tr><th>image</th>
     <td>
      <label id="upload" class="button" style="width: 62px">
       <input type="file" name="image" accept=".gif,.jpg,.jpeg,.png,image/gif,image/jpeg,image/png">
       <span class="rel" style="left: 5px">browse</span>
      </label><input type="text" name="alt" style="width: 314px" placeholder="alt text">
     <td>
    </tr>
   </table>
  </form>
 </header>
 {{range .Previews -}}
 <div class="thread">
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
      {{- if .Op.Tripcode -}}
      {{- if .Op.Verified -}}
      <span class="verif">
      {{- else -}}
      <span class="unverif">
      {{- end -}}
      {{.Op.Tripcode}}</span>
      {{- end -}}
     </span>
     <span class="date">{{.Op.Date}}</span>
     <span class="id">No. {{.Op.Id}}</span>
     [<a href="/{{$.Board}}/{{.Op.Id}}">Reply</a>]
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
  {{range .Replies -}}
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
      {{- if .Tripcode -}}
      {{- if .Verified -}}
      <span class="verif">
      {{- else -}}
      <span class="unverif">
      {{- end -}}
      {{.Tripcode}}</span>
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
 </div>
 {{- end}}
 <div>
  <footer>
   <ul>
   {{range .Pages -}}
   <li>[<a href="/{{$.Board}}/page/{{.}}">{{.}}</a>]</li>
   {{- end}}
   </ul>
  </footer>
 </div>
</body>`
