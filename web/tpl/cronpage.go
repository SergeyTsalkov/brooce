package tpl

var cronPageTpl = `
{{ define "cronitem" }}
<tr {{ if .Disabled }}class="danger"{{ end }}>
  <td class="nowrap">{{ .Name }}</td>
  <td class="wrap"><code>{{ .Raw }}</code></td>

  <td class="buttons">

    <!--
    <a class="btn btn-info btn-xs" href="{{BasePath}}/cron?edit={{ .Name }}">
      <span class="glyphicon glyphicon-edit"></span>
      Edit
    </a>
    -->

    <form action="" method="post">
      <input type="hidden" name="csrf" value="{{CSRF}}">
      <input type="hidden" name="item" value="{{.Name}}">

      {{ if .Disabled }}
        <button type="submit" formaction="{{BasePath}}/enablecron" class="btn btn-warning btn-xs">
          <span class="glyphicon glyphicon-plus"></span>
          Enable
        </button>
      {{ else }}
        <button type="submit" formaction="{{BasePath}}/schedulecron" class="btn btn-success btn-xs">
          <span class="glyphicon glyphicon-repeat"></span>
          Enqueue Now
        </button>      
        <button type="submit" formaction="{{BasePath}}/disablecron" class="btn btn-warning btn-xs">
          <span class="glyphicon glyphicon-remove"></span>
          Disable
        </button>
      {{ end }}

      <button type="submit" formaction="{{BasePath}}/deletecron" onclick="return confirm('Delete Cron Job?')" class="btn btn-danger btn-xs">
        <span class="glyphicon glyphicon-trash"></span>
        Delete
      </button>
    </form>
  </td>
</tr>
{{ end }}

{{ define "cronitemedit" }}
<form action="{{BasePath}}/savecron" method="post">
  <input type="hidden" name="csrf" value="{{CSRF}}">
  <input type="hidden" name="redirect" value="{{BasePath}}/cron">
  <input type="hidden" name="oldname" value="{{ .Name }}">
  <input type="hidden" name="disabled" value="{{ .Disabled }}">

  <tr class="info">
    <td>
      <input type="text" class="form-control" name="name" value="{{ .Name }}">
    </td>
    <td>
      <input type="text" class="form-control" name="item" value="{{ .Raw }}">
    </td>

    <td class="buttons">
      <button type="submit" class="btn btn-success btn-sm">
        <span class="glyphicon glyphicon-ok"></span>
        Save
      </button>

      <a class="btn btn-warning btn-sm" href="{{BasePath}}/cron">
        <span class="glyphicon glyphicon-remove"></span>
        Cancel
      </a>
    </td>
  </tr>

</form>
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

            {{ if eq .Edit "" }}
            <!-- <a class="btn btn-success btn-sm" href="{{BasePath}}/cron?new=1">
              <span class="glyphicon glyphicon-plus"></span>
              New
            </button> -->
            {{ end }}
          </th>
        </tr>
      </thead>
      <tbody>
        {{ range .Crons }}
          {{ if eq .Name $.Edit }} 
            {{ template "cronitemedit" . }}
          {{ else }}
            {{ template "cronitem" . }}
          {{ end }}
        {{ end }}

        {{ range .DisabledCrons }}
          {{ if eq .Name $.Edit }} 
            {{ template "cronitemedit" . }}
          {{ else }}
            {{ template "cronitem" . }}
          {{ end }}
        {{ end }}

        {{ if $.New }}
          {{ template "cronitemedit" }}
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
