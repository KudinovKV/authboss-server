<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{template "pagetitle" .}}</title>
    <meta charset="UTF-8">
</head>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap.min.css" />
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/font-awesome/4.3.0/css/font-awesome.min.css" />
<script type="text/javascript" src="https://code.jquery.com/jquery-2.1.3.min.js"></script>
<script type="text/javascript" src="//maxcdn.bootstrapcdn.com/bootstrap/3.3.2/js/bootstrap.min.js"></script>

<style>
    * {
        margin: 0px;
        padding: 0px;
        box-sizing: border-box;
    }
    form {
        position: relative;
        width: 100%;
        margin: 0 auto;
    }
    body, html {
        height: 100%;
        font-family: Poppins-Regular, sans-serif;
        background: #a64bf4;
        background: -webkit-linear-gradient(45deg, #00dbde, #fc00ff);
        background: -o-linear-gradient(45deg, #00dbde, #fc00ff);
        background: -moz-linear-gradient(45deg, #00dbde, #fc00ff);
        background: linear-gradient(45deg, #00dbde, #fc00ff);
    }
    nav {
         background: rgba(0,0,0,.5) !important;
         box-sizing: border-box !important;
         box-shadow: 0 15px 25px rgba(0,0,0,.6) !important;
         border-radius: 10px !important;
    }
    nav a {
      position: relative;
      display: flex;
      justify-content: center;
      padding: 10px 20px;
      color: #03e9f4;
      font-size: 16px;
      text-decoration: none;
      text-transform: uppercase;
      overflow: hidden;
      transition: .5s;
      letter-spacing: 4px
    }
    nav a:hover {
      background: #03e9f4;
      color: #fff;
      border-radius: 5px;
      box-shadow: 0 0 5px #03e9f4,
                  0 0 25px #03e9f4,
                  0 0 50px #03e9f4,
                  0 0 100px #03e9f4;
    }
    input::-webkit-search-decoration,
    input::-webkit-search-cancel-button,
    input::-webkit-search-results-button,
    input::-webkit-search-results-decoration { display: none; }
</style>
<body class="container-fluid" style="padding-top: 15px;">
    <nav class="navbar">
        <div class="container-fluid">
            <div class="navbar-header">
                <a class="navbar-brand" href="/">Authboss Server</a>
            </div>
            <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
                <ul class="nav navbar-nav navbar-right">
                    {{if not .loggedin}}
                        <li><a href="/auth/register">Register</a></li>
                        <li><a href="/auth/recover">Recover</a></li>
                        <li><a href="/auth/login"><i class="fa fa-sign-in"></i>Login</a></li>
                    {{else}}
                        <li class="dropdown">
                            <a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-expanded="false">Welcome {{.current_user_name}}! <span class="caret"></span></a>
                            <ul class="dropdown-menu" role="menu">
                                <li>
                                    <a href="/auth/logout">
                                        <i class="fa fa-sign-out"></i> Logout
                                    </a>
                                </li>
                            </ul>
                        </li>
                    {{end}}
                </ul>
            </div>
        </div>
    </nav>
	{{with .flash_success}}<div class="alert alert-success">{{.}}</div>{{end}}
	{{with .flash_error}}<div class="alert alert-danger">{{.}}</div>{{end}}
	{{template "yield" .}}
</body>
</html>
{{define "pagetitle"}}{{end}}
{{define "yield"}}{{end}}