package tables

import (
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/html"

	tmpl "html/template"
)

// generators is a map of table models.
//
// The key of generators is the prefix of table info url.
// The corresponding value is the Form and TableName data.
//
// http://{{config.Domain}}:{{Port}}/{{config.Prefix}}/info/{{key}}
//
// example:
//
// "users"   => http://localhost:9033/admin/info/users
// "posts"   => http://localhost:9033/admin/info/posts
// "authors" => http://localhost:9033/admin/info/authors

func label() types.LabelAttribute {
	return template.Get(config.GetTheme()).Label().SetType("success")
}

func interfaces(arr []string) []interface{} {
	var iarr = make([]interface{}, len(arr))

	for key, v := range arr {
		iarr[key] = v
	}

	return iarr
}

func link(url, content string) tmpl.HTML {
	return html.AEl().
		SetAttr("href", url).
		SetContent(template.HTML(content)).
		Get()
}
