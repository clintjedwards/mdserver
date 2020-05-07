package mdserver

import "html/template"

const pageTemplate = `
<!doctype html>

<head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="./{{.Style}}.css">
    <link rel="stylesheet" href="./highlight.css">
    <script src="./highlight.min.js"></script>
    <script>hljs.initHighlightingOnLoad();</script>
</head>

<body>
    <article>
        {{.Body}}
    </article>
</body>
`

var compiledTemplate = template.Must(template.New("page").Parse(pageTemplate))
