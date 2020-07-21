{{define "pagetitle" }}{{.resource}}{{end}}
<style>
.res-box {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 400px;
  padding: 40px;
  transform: translate(-50%, -50%);
  background: rgba(0,0,0,.5);
  box-sizing: border-box;
  box-shadow: 0 15px 25px rgba(0,0,0,.6);
  border-radius: 10px;
}
.res-box h2 {
  margin: 0 0 30px;
  padding: 0;
  color: #fff;
  text-align: center;
}
.res-box .user-box {
  position: relative;
}
.res-box a {
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
  margin-top: 10px;
  letter-spacing: 4px
}
.res-box a:hover {
  background: #03e9f4;
  color: #fff;
  border-radius: 5px;
  box-shadow: 0 0 5px #03e9f4,
              0 0 25px #03e9f4,
              0 0 50px #03e9f4,
              0 0 100px #03e9f4;
}
</style>
{{if .loggedin}}
<div class="res-box">
  <h2>{{.resource}}</h2>
    <div class="user-box">
      <a href="/foo">Foo</a>
      <a href="/bar">Bar</a>
      <a href="/sigma">Sigma</a>
    </div>
</div>
{{end}}