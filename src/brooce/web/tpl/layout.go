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

    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">

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
        margin-top: 3em;
        text-align: center;
      }
      div.pages i {
        display: block;
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

      <li {{ if eq . "overview" }}class="active"{{ end }}><a href="/">Overview</a></li>
      <li {{ if eq . "cron" }}class="active"{{ end }}><a href="/cron">Cron Jobs</a></li>
    </ul>

    <h1 class="text-muted"><a href="/">brooce</a></h1>
  </div>
{{ end }}
`

var footerTpl = `
{{ define "footer" }}
</div> <!-- container -->
</body>

<script src="https://code.jquery.com/jquery-2.2.4.min.js" integrity="sha256-BbhdlvQf/xTY9gja0Dq3HiwQF8LaCRTXxZKRutelT44=" crossorigin="anonymous"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js" integrity="sha384-0mSbJDEHialfmuBBQP6A4Qrprq5OVfW37PRR3j5ELqxss1yVqOtnepnHVP9aJ7xS" crossorigin="anonymous"></script>

</html>
{{ end }}
`
