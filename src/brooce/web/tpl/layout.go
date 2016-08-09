package tpl

var headerTpl = `
{{ define "header" }}

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">

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

      <li class="active"><a href="/">Overview</a></li>
      <li><a href="#">Schedule</a></li>
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
