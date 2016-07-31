package tpl

var mainPageTpl = `
{{ define "mainpage" }}
{{ template "header" }}
<div class="row">
  <div class="col-md-6">
    <h3>Queues</h3>
    <table class="table">
      <thead>
        <tr>
          <th>Queue</th>
          <th>Pending</th>
        </tr>
      </thead>
      <tbody>
        {{ range $queueName, $queueLength := .ListQueues }}
          <tr>
            <td>{{ $queueName }}</td>
            <td>{{ $queueLength }}</td>
          </tr>
        {{ end }}
      </tbody>
    </table>

  </div>
</div>


<div class="row">
  <div class="col-md-12">
    <h3>1 Worker Alive</h3>
    <table class="table">
      <thead>
        <tr>
          <th>Worker Name</th>
          <th>Machine Name</th>
          <th>Machine IP</th>
          <th>Process ID</th>
          <th>Workers Active</th>
          <th>Queues</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>69.90.132.223-4135</td>
          <td>sdev</td>
          <td>69.90.132.223</td>
          <td>4135</td>
          <td>2/5</td>
          <td>3x<tt>common</tt>, 2x<tt>special</tt></td>
        </tr>
      </tbody>
    </table>
  </div>
</div>



<div class="row">
  <div class="col-md-12">

    <h3>2 of 5 Threads Working</h3>
    <table class="table">
      <thead>
        <tr>
          <th>Worker Name</th>
          <th>Queue</th>
          <th>Runtime</th>
          <th>Command</th>
          <th>Params</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>69.90.132.223-4135</td>
          <td>common</td>
          <td>2 minutes 8 seconds</td>
          <td><code>rm -fr /</code></td>
          <td><code>{}</code></td>
        </tr>
      </tbody>
    </table>

  </div>
</div>
{{ template "footer" }}
{{ end }}
`
