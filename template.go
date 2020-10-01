package main

import "text/template"

var tmpl = template.Must(template.New("template").Parse(`+++ 
date = "{{.Date}}"
title = "{{.Title}}"
slug = "qiita-{{.ID}}" 
tags = [{{.AllTags}}]
categories = []
+++

{{.Body}}

*この記事は[Qiita]({{.URL}})の記事をエクスポートしたものです*
`))
