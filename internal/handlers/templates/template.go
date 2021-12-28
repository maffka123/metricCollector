package templates

var MetricTemplate = `<html>
    <head>
    <title>(/^▽^)/</title>
    </head>
    <body>
        <h1>Counter</h1>>
    {{range .Counter}}
            <li>{{.NameValue}}</li>
    {{end}}

    <h1>Gauge</h1>>
    {{range .Gauge}}
    <li>{{.NameValue}}</li>
{{end}}

    </body>
</html>`
