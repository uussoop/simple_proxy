package panel

import (
	"html/template"
	"strconv"

	"github.com/GoAdminGroup/go-admin/context"
	tmpl "github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/rodrikv/openai_proxy/database"
	"github.com/sirupsen/logrus"
)

// GetContent return the content of index page.
func DashboardPage(ctx *context.Context) (panel types.Panel, err error) {

	components := tmpl.Default()
	colComp := components.Col()

	/**************************
	 * Box
	/**************************/

	results := []map[string]types.InfoItem{}

	rows, err := database.Db.Raw("SELECT proxy_models.name AS model_name, proxy_users.name AS user_name, SUM(proxy_endpoint_model_usages.token_used) AS total_tokens_used FROM proxy_endpoint_model_usages LEFT JOIN proxy_models ON proxy_endpoint_model_usages.model_id = proxy_models.id LEFT JOIN proxy_users ON proxy_endpoint_model_usages.user_id = proxy_users.id GROUP BY proxy_models.name, proxy_users.name;").Rows()
	if err != nil {
		logrus.Error(err)
	}
	for rows.Next() {
		var model_name string
		var user_name string
		var total_tokens_used uint
		if err := rows.Scan(&model_name, &user_name, &total_tokens_used); err != nil {
			// Handle the error
			continue
		}
		result := map[string]types.InfoItem{
			"User":              {Content: template.HTML(user_name)},
			"Model Name":        {Content: template.HTML(model_name)},
			"Total Tokens Used": {Content: template.HTML(strconv.Itoa(int(total_tokens_used)))},
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return panel, err
	}
	defer rows.Close()

	table := components.Table().SetType("table").SetInfoList(results).SetThead(types.Thead{
		{Head: "User"},
		{Head: "Model Name"},
		{Head: "Total Tokens Used"},
	}).GetContent()

	boxInfo := components.Box().
		WithHeadBorder().
		SetHeader("Total User Usage").
		SetHeadColor("#f7f7f7").
		SetBody(table).
		GetContent()

	tableCol := colComp.SetSize(types.SizeMD(8)).SetContent(boxInfo).GetContent()

	row5 := components.Row().SetContent(tableCol).GetContent()

	return types.Panel{
		Content:     row5,
		Title:       "Dashboard",
		Description: "Usage Shower",
	}, nil
}
