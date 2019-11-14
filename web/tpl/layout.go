package tpl

var headerTpl = `
{{ define "header" }}

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <link id="favicon" rel="shortcut icon" type="image/png" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAABsElEQVR42mNgoBZwc3PjTkxMlC0oKFDAhjMyMqRDQ0M5sWr28/Pj9fDwSAfiXQEBAcewYS8vry2urq75CQkJChgGADUq2NjYHFNSUvqnrKz8TUVF5SMyBop9BMp9NTQ0/Obr61teWFjIiWGAubn5MXV19W86OjpzgXQiMlZVVQXhSUDD/puYmEwHekcMqwHa2tofgXQiugv379/PAfRGPsgAoPzM/Px8cXwGJOnp6XEbGBgIgLC7u7tQaWlpOJD9EGjAH1NT05r58+dz4DTAzMysCujfzUD8AYaBYfAVFD5A598IDw/3xBqI+AwAYUVFxS9A+q+dnd255ORkM5wGWFpaJmdlZfFkZmYKwnBUVJSgvb29KzAWrgNd89/R0bEIJSYIBSII1NfXMwHDoURNTe0/MMp7gelBAMMAoORXTU3NOfLy8onoGOj8JC0trRXA6PxvYWHREh0dzQc3ABhFCkC/gRLSX6BfvwLxB3QMlAMlqp9ACx4C1QalpaWxwg1ISkridXJyigEq6lNQUMCJgYmqw9raOhyYJ4Qw/Ojg4MAhJycniA8DLeAHeoONgZoAAFeoynJXT6jnAAAAAElFTkSuQmCC">

    <title>Brooce Job Queue</title>

    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/3.4.1/css/bootstrap.min.css" integrity="sha256-bZLfwXAP04zRMK2BjiO8iu9pf4FbLqX6zitd+tIvLhE=" crossorigin="anonymous" />

    <style>
      div.header {
        margin: 0.5em 0 1em 0;
        padding: 0.5em 0 0.5em 0;
        border-bottom: 1px solid #777;
      }
      div.header h1 {
        font-size: 2em;
        font-weight: bold;
        margin: 0;
        line-height: 40px;
      }
      div.header h1 a {
        text-decoration: none;
        color: inherit;
      }
      div.header h1 a:hover {
        text-decoration: none;
      }

      div.pages {
        margin: 3em 0 1em 0;
        text-align: center;
      }
      div.pages i {
        display: block;
      }
      div.pages .prevnext {
        font-weight: bold;
        margin: 0 1em;
      }

      td.buttons, th.buttons {
        text-align: right;
      }

      td.buttons form, th.buttons form {
        display: inline-block;
        margin-left: 1em;
      }

      pre {
        width: 100%;
        height: 90vh;
      }

      td.params {
        font-family: monospace;
        font-size: 0.75em;
      }

      td.params ul {
        list-style: none;
        margin: 0;
        padding: 0;
      }

      td.wrap {
        word-break: break-all;
      }

      td.nowrap {
        white-space: nowrap;
      }

      th.queuename,td.queuename {
        text-align: right;
      }

      th.queuenum,td.queuenum {
        text-align: center;
      }


    </style>


    <!-- HTML5 shim and Respond.js for IE8 support of HTML5 elements and media queries -->
    <!-- WARNING: Respond.js doesn't work if you view the page via file:// -->
    <!--[if lt IE 9]>
      <script src="https://oss.maxcdn.com/html5shiv/3.7.2/html5shiv.min.js"></script>
      <script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
    <![endif]-->

  </head>
  <body>

  <div class="container">
  <div class="header clearfix">
    <ul class="nav nav-pills pull-right">

      <li {{ if eq . "overview" }}class="active"{{ end }}><a href="{{BasePath}}/" accesskey="o" title="Alt+O">Overview</a></li>
      <li {{ if eq . "cron" }}class="active"{{ end }}><a href="{{BasePath}}/cron" accesskey="j" title="Alt+J">Cron Jobs</a></li>
    </ul>

    <h1 class="text-muted"><a href="{{BasePath}}/">brooce</a></h1>
  </div>
{{ end }}
`

var footerTpl = `
{{ define "footer" }}
</div> <!-- container -->
</body>

{{ if . }}
{{ if eq . "joblist" }}
  <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.2.4/jquery.min.js" integrity="sha256-BbhdlvQf/xTY9gja0Dq3HiwQF8LaCRTXxZKRutelT44=" crossorigin="anonymous"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/js-cookie/2.2.0/js.cookie.min.js" integrity="sha256-9Nt2r+tJnSd2A2CRUvnjgsD+ES1ExvjbjBNqidm9doI=" crossorigin="anonymous"></script>
  <script>
    $(document).keydown(function(evt) {
      var target = null
      if (evt.which == 37) target = $('a.prev-link')
      else if (evt.which == 39) target = $('a.next-link')
      else return

      if (target && target.length > 0) {
        window.location.href = target.attr('href')
      }
    })

    $('a.setresults-link').click(function(evt) {
      evt.preventDefault()
      var perpage = prompt("How may results per page? (1-100, default 10)")
      Cookies.set('perpage', parseInt(perpage))
      location.reload()
    })
  </script>
{{ end }}
{{ end }}

<!--
<script src="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/3.4.1/js/bootstrap.min.js" integrity="sha256-nuL8/2cJ5NDSSwnKD8VqreErSWHtnEP9E7AySL+1ev4=" crossorigin="anonymous"></script>
-->

</html>
{{ end }}
`
