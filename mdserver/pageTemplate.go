package mdserver

import "html/template"

const pageTemplate = `
<!doctype html>

<head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
		{{.Style}}
    </style>
</head>

<body>
    <article>
        {{.Body}}
    </article>
</body>
`

var compiledTemplate = template.Must(template.New("page").Parse(pageTemplate))
