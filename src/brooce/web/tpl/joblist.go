package tpl

var jobListTpl = `
{{ define "joblist" }}
{{ template "header" }}

<div class="row">
  <div class="col-md-12">

    <h3>{{ .ListType }} jobs for queue {{ .QueueName }}</h3>
    <table class="table">
      <thead>
        <tr>
          <th>Command</th>
          <th>Params</th>
        </tr>
      </thead>
      <tbody>
        {{ range .Jobs }}
          <tr>
            <td><code>{{ .FullCommand }}</code></td>
            <td><code></code></td>
          </tr>
        {{ end }}
      </tbody>
    </table>
    
    <div class="pages">
      <i>Showing results {{ .Start }}-{{ .End }} of {{ .Length }}</i>

      {{ if gt .Pages 1 }}
        {{ range Iter 1 .Pages }}
          {{ if eq $.Page . }}
            {{ . }}
          {{ else }}
            <a href="?{{ . }}">{{ . }}</a>
          {{ end }}
        {{ end }}
      {{ end }}
    </div>
    
  </div>
</div>
{{ template "footer" }}
{{ end }}
`
