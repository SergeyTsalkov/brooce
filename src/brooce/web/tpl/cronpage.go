package tpl

var cronPageTpl = `
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
          <th></th>
        </tr>
      </thead>
      <tbody>
        {{ range . }}
          <tr>
            <td>{{ .Name }}</td>
            <td><tt>{{ .Raw }}</tt></td>

            <td class="buttons">
              <button class="btn btn-info btn-xs">
                <span class="glyphicon glyphicon-edit"></span>
                Edit
              </button>

              <button class="btn btn-warning btn-xs">
                <span class="glyphicon glyphicon-remove"></span>
                Disable
              </button>

              <button class="btn btn-danger btn-xs">
                <span class="glyphicon glyphicon-trash"></span>
                Delete
              </button>
            </td>
          </tr>
        {{ end }}
      </tbody>
    </table>
    
    
  </div>
</div>
{{ template "footer" }}
{{ end }}
`
