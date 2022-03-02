// holds htm template to be returned by one of the endpoints.
// I just did not find how to properly set path to .html files ^^
package templates

// MetricTemplate string that holds html-template with musters in {{ brackets.
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
