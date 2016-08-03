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
        {{ range $i, $Job := .Jobs }}
          <tr>
            <td><code>{{ $Job.FullCommand }}</code></td>
            <td><code></code></td>
          </tr>
        {{ end }}
      </tbody>
    </table>

  </div>
</div>
{{ template "footer" }}
{{ end }}
`
