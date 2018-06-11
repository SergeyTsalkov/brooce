package tpl

var jobListTpl = `
{{ define "joblist" }}
{{ template "header" "" }}

<div class="row">
  <div class="col-md-12">

    <div class='row'>
      <div class='col-sm-3'>
        <form class="form-search" method="get" action="/search">
          <input type="hidden" name="queue" value="{{ .QueueName }}">
          <input type="hidden" name="listType" value="{{ .ListType }}">
          <div class="input-group">
          <input name="q" type="text" class="form-control search-query" placeholder="Search by Command" value="{{ .Query }}">
            <div class="input-group-btn">
              <button class="btn btn-default" type="submit"><i class="glyphicon glyphicon-search"></i></button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <h3>{{ .ListType }} jobs for queue {{ .QueueName }}</h3>
    <table class="table table-hover">
      <thead>
        <tr>
          {{ if eq .ListType "done" "failed" }}
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
                <span class="glyphicon glyphicon-trash"></span>
                Delete All
              </button>
            </form>
          </th>
        </tr>
      </thead>
      <tbody>
        {{ range .Jobs }}
          <tr>
            {{ if eq $.ListType "done" "failed" }}
              <td><span title="{{FormatTime .EndTime}}">{{ TimeSince .EndTime }}</span></td>
              <td>{{ TimeBetween .EndTime .StartTime }}</td>
            {{ end }}
            <td><code>{{ .Command }}</code></td>
            <td class="params">
              <ul>
                {{ if .Timeout }} <li>Timeout: {{ TimeDuration .Timeout }} {{ end }}
                {{ if gt .MaxTries 1 }} <li>Max Tries: {{ .MaxTries }} {{ end }}
                {{ if .Cron }} <li>Cron: {{ .Cron }} {{ end }}
                {{ if .Locks }} <li>Locks: {{ Join .Locks ", " }} {{ end }}
              </ul>
            </td>
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
                  <span class="glyphicon glyphicon-trash"></span>
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
      
      {{ if lt $.Page 2 }}
        <span class="prevnext">&#10235; Prev</span>
      {{ else }}
        <a class="prevnext" href="?{{ $.LinkParamsForPrevPage $.Page}}">&#10235; Prev</a>
      {{ end }}

      Page {{ $.Page }} of {{ .Pages }}

      {{ if eq $.Page $.Pages }}
        <span class="prevnext">Next &#10236;</span>
      {{ else }}
        <a class="prevnext" href="?{{ $.LinkParamsForNextPage $.Page}}">Next &#10236;</a>
      {{ end }}
    </div>
    
  </div>
</div>
{{ template "footer" }}
{{ end }}
`
