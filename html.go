package main

import (
	"io/ioutil"
	"log"
)

const (
	htmlMaster = `
	{{define "master"}}
	<html>
	<head>
		<title>GJFY - {{template "title" .}}</title>
		<link rel="shortcut icon" type="image/x-icon" href="favicon.ico" />
		<link rel="stylesheet" type="text/css" href="custom.css">
	</head>
	<body>
	<div id="contentcontainer">
	<div id="content">
	{{template "content" .}}
	</div>
	</div>
	</body>
	</html>
	{{end}}
	`
	htmlView = `
	{{define "title"}}VIEW{{end}}
	{{define "content"}}
	<h2 id="mainheading">{{.Id}}</h2>
	<div>
		The link you just invoked contains a secret (e. g. a password) somebody wants to share with you.
		It will be valid only for a short time and you might not be able to invoke it again.
		Please make sure you memorise the secret or write it down in an appropriate way.
	</div>
	<div>The secret contained in this link is as follows:</div>
	<div id="secret">{{.Secret}}</div>
	{{end}}
	`
	htmlViewInfo = `
	{{define "title"}}VIEWINFO{{end}}
	{{define "content"}}
	<h2 id="mainheading">{{.Id}}</h2>
	<table id="info">
	<tr>
		<th>Url</th>
		<th>PathQuery</th>
		<th>MaxClicks</th>
		<th>Clicks</th>
		<th>DateAdded</th>
	</tr>
	<tr>
		<td><a href="{{.Url}}">{{.Url}}</a></td>
		<td>{{.PathQuery}}</td>
		<td>{{.MaxClicks}}</td>
		<td>{{.Clicks}}</td>
		<td>{{.DateAdded}}</td>
	</tr>
	</table>
	{{end}}
	`
	htmlViewErr = `
	{{define "title"}}ERROR{{end}}
	{{define "content"}}
	<h2 id="errorheading">Not available</h2>
	<p id="errormessage">This ID is not valid anymore. Please request another one from the person who sent you this link.</p>
	{{end}}
	`
)

var (
	favicon = []byte{0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x10, 0x10, 0x10, 0x0, 0x1,
		0x0, 0x4, 0x0, 0x28, 0x1, 0x0, 0x0, 0x16, 0x0, 0x0, 0x0, 0x28, 0x0, 0x0,
		0x0, 0x10, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0, 0x1, 0x0, 0x4, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff, 0x0, 0x0, 0xff,
		0xff, 0x0, 0x0, 0xfe, 0x7f, 0x0, 0x0, 0xfe, 0x7f, 0x0, 0x0, 0xff, 0xff,
		0x0, 0x0, 0xfe, 0x7f, 0x0, 0x0, 0xfe, 0x7f, 0x0, 0x0, 0xff, 0x3f, 0x0, 0x0,
		0xff, 0x9f, 0x0, 0x0, 0xfd, 0xdf, 0x0, 0x0, 0xfc, 0x9f, 0x0, 0x0, 0xfe,
		0x3f, 0x0, 0x0, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff,
		0x0, 0x0}
	cssFile = "/etc/gjfy/custom.css"
)

func readCssFile() []byte {
	css, err := ioutil.ReadFile(cssFile)
	if err != nil {
		log.Println("could not read css file from", cssFile)
		css = []byte{}
	}
	return css
}
