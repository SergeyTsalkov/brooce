package tpl

var showLogTpl = `
{{ define "showlog" }}
{{ template "header" "" }}

<div class="row">
  <div class="col-md-12">
    <pre>{{.}}</pre>
  </div>
</div>
{{ template "footer" }}
{{ end }}
`
