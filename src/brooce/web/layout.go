package web

var tplList = []string{
	headerTpl,
	footerTpl,
	mainPageTpl,
}

var headerTpl = `
{{ define "header" }}
<html>
  <head>
    <title>Brooce Control Panel</title>
  </head>
  <body>
{{ end }}
`

var footerTpl = `
{{ define "footer" }}
  </body>
</html>
{{ end }}
`

var mainPageTpl = `
{{ template "header" }}
new template!
{{ template "footer" }}
`
