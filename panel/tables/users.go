package tables

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	editType "github.com/GoAdminGroup/go-admin/template/types/table"
	"github.com/rodrikv/openai_proxy/pkg/token"
	"github.com/sirupsen/logrus"
)

// GetUserTable return the model of table user.
func (s *SystemTable) GetUserTable(ctx *context.Context) (userTable table.Table) {

	userTable = table.NewDefaultTable(table.Config{
		Driver:     db.DriverSqlite,
		CanAdd:     true,
		Editable:   true,
		Deletable:  true,
		Exportable: true,
		Connection: table.DefaultConnectionName,
		PrimaryKey: table.PrimaryKey{
			Type: db.Int,
			Name: table.DefaultPrimaryKeyName,
		},
	})

	info := userTable.GetInfo().SetFilterFormLayout(form.LayoutThreeCol)
	{
		info.AddField("ID", "id", db.Int).FieldSortable()
		info.AddField("Name", "name", db.Varchar).FieldEditAble(editType.Text).
			FieldFilterable(types.FilterType{Operator: types.FilterOperatorLike})
		info.AddField("Token", "token", db.Varchar).FieldEditAble(editType.Text).FieldFilterable()
		info.AddField("RequestCount", "request_count", db.Int).FieldFilterable()
		info.AddField("RateLimit", "rate_limit", db.Int).FieldFilterable()
		info.AddField("TokenLimit", "token_limit", db.Int).FieldFilterable()
		info.AddField("UsageToday", "usage_today", db.Int).FieldFilterable()
		info.AddField("Limited", "limited", db.Bool)
		info.AddField("CreatedAt", "created_at", db.Timestamp).
			FieldFilterable(types.FilterType{FormType: form.DatetimeRange})
		info.AddField("UpdatedAt", "updated_at", db.Timestamp).FieldEditAble(editType.Datetime)
		info.AddField("DeletedAt", "deleted_at", db.Timestamp).
			FieldFilterable(types.FilterType{FormType: form.DatetimeRange})
		info.AddField("LastSeen", "last_seen", db.Timestamp).
			FieldFilterable(types.FilterType{FormType: form.DatetimeRange})

		info.AddField("Endpoints", "name", db.Text).
			FieldJoin(types.Join{
				Table:     "proxy_user_endpoints",
				JoinField: "user_id",
				Field:     "id",
			}).
			FieldJoin(types.Join{
				Table:     "proxy_endpoints",
				JoinField: "id",
				Field:     "endpoint_id",
				BaseTable: "proxy_user_endpoints",
			}).
			FieldDisplay(func(model types.FieldModel) interface{} {
				fmt.Print(model.Value)
				labels := template.HTML("")
				labelTpl := label().SetType("success")

				labelValues := strings.Split(model.Value, types.JoinFieldValueDelimiter)
				for key, label := range labelValues {
					if key == len(labelValues)-1 {
						labels += labelTpl.SetContent(template.HTML(label)).GetContent()
					} else {
						labels += labelTpl.SetContent(template.HTML(label)).GetContent() + ""
					}
				}

				if labels == template.HTML("") {
					return "no endpoints"
				}

				return labels
			}).FieldFilterable()

		info.SetDeleteFn(func(idArr []string) (err error) {
			var ids = interfaces(idArr)

			_, err = s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteUserRoleErr := s.connection().WithTx(tx).
					Table("proxy_user_endpoints").
					WhereIn("user_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}
				deleteUserRoleErr = s.connection().WithTx(tx).
					Table("proxy_users").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}

				return nil, nil
			})
			return
		})

		info.SetTable("proxy_users").SetTitle("Users").SetDescription("Users")
	}
	formList := userTable.GetForm()

	{
		formList.AddField("Name", "name", db.Varchar, form.Text).FieldMust()
		formList.AddField("Token", "token", db.Varchar, form.Text).FieldMust().FieldDefault("sk-" + token.String(32)).FieldPlaceholder("Token")
		formList.AddField("RequestCount", "request_count", db.Int, form.Number).FieldDefault("0")
		formList.AddField("RateLimit", "rate_limit", db.Int, form.Number).FieldMust().FieldDefault("10")
		formList.AddField("TokenLimit", "token_limit", db.Int, form.Number).FieldMust().FieldDefault("40000")
		formList.AddField("UsageToday", "usage_today", db.Int, form.Number).FieldMust().FieldDefault("0")

		formList.AddField("Endpoints", "endpoints", db.Varchar, form.Select).
			FieldOptionsFromTable("proxy_endpoints", "name", "id").
			FieldDisplay(func(model types.FieldModel) interface{} {
				var endpoints []string

				if model.ID == "" {
					return endpoints
				}
				permissionModel, _ := s.table("proxy_user_endpoints").
					Select("endpoint_id").Where("user_id", "=", model.ID).All()
				for _, v := range permissionModel {
					endpoints = append(endpoints, strconv.FormatInt(v["endpoint_id"].(int64), 10))
				}
				return endpoints
			}).FieldHelpMsg(template.HTML("no corresponding options?") +
			link("/admin/info/endpoints/new", "Create here."))

		formList.AddField("CreatedAt", "created_at", db.Timestamp, form.Datetime).FieldNotAllowEdit().FieldDefault(time.Now().String())

		formList.SetTable("proxy_users").SetTitle("Users").SetDescription("Users")
		formList.SetInsertFn(
			func(values form2.Values) (err error) {
				_, err = s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {
					modelId, insertUserRoleErr := s.connection().WithTx(tx).
						Table("proxy_users").
						Insert(dialect.H{
							"name":        values.Get("name"),
							"token":       values.Get("token"),
							"rate_limit":  values.Get("rate_limit"),
							"token_limit": values.Get("token_limit"),
							"usage_today": values.Get("usage_today"),
						})

					if db.CheckError(insertUserRoleErr, db.INSERT) {
						return insertUserRoleErr, nil
					}
					subModelsIds := values["endpoints[]"]
					logrus.Info(subModelsIds)
					for _, v := range subModelsIds {
						_, insertUserRoleErr = s.connection().WithTx(tx).
							Table("proxy_user_endpoints").
							Insert(dialect.H{
								"user_id":     modelId,
								"endpoint_id": v,
								"updated_at":  time.Now().String(),
								"created_at":  time.Now().String(),
							})

						if db.CheckError(insertUserRoleErr, db.INSERT) {
							return insertUserRoleErr, nil
						}
					}
					return nil, nil
				})
				return
			},
		)

		formList.SetUpdateFn(
			func(values form2.Values) (err error) {
				_, err = s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {
					modelId := values.Get("id")

					_, insertUserRoleErr := s.connection().WithTx(tx).
						Table("proxy_users").
						Where("id", "=", modelId).
						Update(dialect.H{
							"name":          values.Get("name"),
							"token":         values.Get("token"),
							"request_count": values.Get("request_count"),
							"rate_limit":    values.Get("rate_limit"),
							"token_limit":   values.Get("token_limit"),
							"usage_today":   values.Get("usage_today"),
						})

					if db.CheckError(insertUserRoleErr, db.UPDATE) {
						return errors.New("update: " + insertUserRoleErr.Error()), nil
					}

					_ = s.connection().WithTx(tx).
						Table("proxy_user_endpoints").
						Where("user_id", "=", modelId).
						Delete()

					subModelsIds := values["endpoints[]"]
					logrus.Info(subModelsIds)
					for _, v := range subModelsIds {
						_, insertUserRoleErr = s.connection().WithTx(tx).
							Table("proxy_user_endpoints").
							Insert(dialect.H{
								"user_id":     modelId,
								"endpoint_id": v,
								"updated_at":  time.Now().String(),
								"created_at":  time.Now().String(),
							})

						if db.CheckError(insertUserRoleErr, db.INSERT) {
							return insertUserRoleErr, nil
						}
					}
					return nil, nil
				})

				return
			},
		)

		formList.SetPostHook(func(values form2.Values) error {
			fmt.Println("userTable.GetForm().PostHook", values)
			return nil
		})
	}

	userTable.GetForm().SetTabGroups(types.
		NewTabGroups("name", "token", "rate_limit", "request_count", "token_limit", "usage_today", "limited", "endpoints", "created_at")).
		// AddGroup("phone", "role", "request_count", "updated_at")).
		SetTabHeaders("New User")

	return
}
