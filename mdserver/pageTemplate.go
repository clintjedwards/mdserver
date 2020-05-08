package mdserver

import "html/template"

const pageTemplate = `
<!doctype html>

<head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/css/{{.Style}}.css">
    <link rel="stylesheet" href="/css/highlight.css">
    <script src="/javascript/highlight.min.js"></script>
    <script>hljs.initHighlightingOnLoad();</script>
</head>

<body>
    <article>
        {{.Body}}
    </article>
</body>
</html>
`

var compiledTemplate = template.Must(template.New("page").Parse(pageTemplate))
