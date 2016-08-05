package tpl

var jobListTpl = `
{{ define "joblist" }}
{{ template "header" }}

<div class="row">
  <div class="col-md-12">

    <h3>{{ .ListType }} jobs for queue {{ .QueueName }}</h3>
    <table class="table table-hover">
      <thead>
        <tr>
          {{ if ne .ListType "pending" }}
            <th>Finished</th>
            <th>Runtime</th>
          {{ end }}

          <th>Command</th>
          <th>Params</th>
          <th class="buttons">
            <form action="" method="post">
              <input type="hidden" name="csrf" value="{{CSRF}}">
              {{ if eq .ListType "failed" "delayed" }}
              <button type="submit" formaction="/retryall/{{ .ListType }}/{{ .QueueName }}" class="btn btn-warning btn-sm">
                <span class="glyphicon glyphicon-repeat"></span>
                Retry All
              </button>
              {{ end }}

              <button type="submit" formaction="/deleteall/{{ .ListType }}/{{ .QueueName }}" class="btn btn-danger btn-sm">
                <span class="glyphicon glyphicon-remove"></span>
                Delete All
              </button>
            </form>
          </th>
        </tr>
      </thead>
      <tbody>
        {{ range .Jobs }}
          <tr>
            {{ if ne $.ListType "pending" }}
              <td><span title="{{FormatTime .EndTime}}">{{ TimeSince .EndTime }} ago</span></td>
              <td>{{ TimeBetween .EndTime .StartTime }}</td>
            {{ end }}
            <td><code>{{ .FullCommand }}</code></td>
            <td><code></code></td>
            <td class="buttons">
              {{ if .HasLog }}
                <a href="/showlog/{{ .Id }}" target="_new" class="btn btn-info btn-xs">
                  <span class="glyphicon glyphicon-align-justify"></span>
                  Show Log
                </a>
              {{ end }}

              <form action="" method="post">
                <input type="hidden" name="csrf" value="{{CSRF}}">
                <input type="hidden" name="item" value="{{.Raw}}">

                {{ if eq $.ListType "failed" "delayed" "done" }}
                <button type="submit" formaction="/retry/{{ $.ListType }}/{{ $.QueueName }}" class="btn btn-warning btn-xs">
                  <span class="glyphicon glyphicon-repeat"></span>
                  Retry
                </button>
                {{ end }}


                <button type="submit" formaction="/delete/{{ $.ListType }}/{{ $.QueueName }}" class="btn btn-danger btn-xs">
                  <span class="glyphicon glyphicon-remove"></span>
                  Delete
                </button>
              </form>
            </td>
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
