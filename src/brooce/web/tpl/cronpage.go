package tpl

var cronPageTpl = `
{{ define "cronitem" }}
<tr {{ if .Disabled }}class="danger"{{ end }}>
  <td>{{ .Name }}</td>
  <td><tt>{{ .Raw }}</tt></td>

  <td class="buttons">
    <button class="btn btn-info btn-xs">
      <span class="glyphicon glyphicon-edit"></span>
      Edit
    </button>

    <form action="" method="post">
      <input type="hidden" name="csrf" value="{{CSRF}}">
      <input type="hidden" name="item" value="{{.Name}}">

      {{ if .Disabled }}
        <button type="submit" formaction="/enablecron" class="btn btn-warning btn-xs">
          <span class="glyphicon glyphicon-plus"></span>
          Enable
        </button>
      {{ else }}
        <button type="submit" formaction="/disablecron" class="btn btn-warning btn-xs">
          <span class="glyphicon glyphicon-remove"></span>
          Disable
        </button>
      {{ end }}

      <button type="submit" formaction="/deletecron" onclick="return confirm('Delete Cron Job?')" class="btn btn-danger btn-xs">
        <span class="glyphicon glyphicon-trash"></span>
        Delete
      </button>
    </form>
  </td>
</tr>
{{ end }}

{{ define "cronpage" }}
{{ template "header" "cron" }}
<div class="row">
  <div class="col-md-12">

    <h3>Cron Jobs</h3>
    <table class="table table-hover">
      <thead>
        <tr>
          <th>Name</th>
          <th>Job</th>
          <th class="buttons">
            <button class="btn btn-success btn-sm">
              <span class="glyphicon glyphicon-plus"></span>
              New
            </button>
          </th>
        </tr>
      </thead>
      <tbody>
        {{ range .Crons }}
          {{ template "cronitem" . }}
        {{ end }}

        {{ range .DisabledCrons }}
          {{ template "cronitem" . }}
        {{ end }}
      </tbody>
    </table>

    <center>
      <i>current UTC time is {{ CurrentTime }}</i>
    </center>
    
    
  </div>
</div>
{{ template "footer" }}
{{ end }}
`
