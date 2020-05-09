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
	<h1>Test Index Header</h1>
	<input type="text" placeholder="Search.." name="search">
</body>
</html>
`

var compiledIndexTemplate = template.Must(template.New("index").Parse(indexTemplate))
