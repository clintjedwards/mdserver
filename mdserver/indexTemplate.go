package mdserver

import "html/template"

const indexTemplate = `
<!doctype html>

<head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/css/{{.Style}}.css">
</head>

<body>
	<h1>Markdown Server</h1>
	<input type="text" placeholder="Search" name="search" autocomplete="off" />
	<br />

	<table class="center">
	<tr>
	  <th>Name</th>
	  <th>Last Modified</th>
	  <th>Size</th>
	</tr>
	{{range .Files}}
	<tr>
		<td style="text-transform:capitalize;"><a href="{{.Path}}">{{.Name}}</a></td>
		<td><a href="{{.Path}}">{{.Modified}}</a></td>
		<td><a href="{{.Path}}">{{.Size}}</a></td>
	</tr>
	{{end}}
  </table>
</body>
</html>
`

var compiledIndexTemplate = template.Must(template.New("index").Parse(indexTemplate))
