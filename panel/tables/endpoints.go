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
	"github.com/sirupsen/logrus"
)

// GetUserTable return the model of table user.
func (s *SystemTable) GetEndpointTable(ctx *context.Context) (userTable table.Table) {

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

	// type Endpoint struct {
	// 	gorm.Model
	// 	Name         string `json:"name"`
	// 	Url          string `json:"url"`
	// 	Token        string `json:"token"`
	// 	Concurrent   int    `json:"concurrent" gorm:"default:4"`
	// 	Connections  int    `json:"connections"`
	// 	IsActive     bool   `json:"is_active" gorm:"default:true"`
	// 	RPM          int    `json:"rpm" gorm:"type:int;default:3"`
	// 	RPD          int    `json:"rpd" gorm:"type:int;default:200"`
	// 	RequestInMin int    `json:"request_in_min" gorm:"type:int;default:0"`
	// 	RequestInDay int    `json:"request_in_day" gorm:"type:int;default:0"`

	// 	Models []Model `gorm:"many2many:endpoint_models;"`
	// 	Users  []User  `gorm:"many2many:user_endpoints;"`
	// }

	info := userTable.GetInfo().SetFilterFormLayout(form.LayoutThreeCol)
	{
		info.AddField("ID", "id", db.Int).FieldSortable()
		info.AddField("Name", "name", db.Varchar).FieldEditAble(editType.Text).
			FieldFilterable(types.FilterType{Operator: types.FilterOperatorLike})
		info.AddField("Url", "url", db.Varchar).FieldEditAble(editType.Text).FieldFilterable()
		info.AddField("Token", "token", db.Varchar).FieldFilterable()
		info.AddField("Concurrent", "concurrent", db.Int).FieldFilterable()
		info.AddField("Connections", "connections", db.Int).FieldFilterable()
		info.AddField("IsActive", "is_active", db.Bool)
		info.AddField("RPM", "rpm", db.Int).FieldFilterable()
		info.AddField("RPD", "rpd", db.Int).FieldFilterable()
		info.AddField("RequestInMin", "request_in_min", db.Int).FieldFilterable()
		info.AddField("RequestInDay", "request_in_day", db.Int).FieldFilterable()

		info.AddField("Models", "name", db.Varchar).
			FieldJoin(types.Join{
				Table:     "proxy_endpoint_models",
				JoinField: "endpoint_id",
				Field:     "id",
			}).
			FieldJoin(types.Join{
				Table:     "proxy_models",
				JoinField: "id",
				Field:     "model_id",
				BaseTable: "proxy_endpoint_models",
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
					return "no models"
				}

				return labels
			}).FieldFilterable()

		info.AddField("CreatedAt", "created_at", db.Timestamp).
			FieldFilterable(types.FilterType{FormType: form.DatetimeRange})
		info.AddField("UpdatedAt", "updated_at", db.Timestamp).FieldEditAble(editType.Datetime)
		info.AddField("DeletedAt", "deleted_at", db.Timestamp).
			FieldFilterable(types.FilterType{FormType: form.DatetimeRange})

		info.SetDeleteFn(func(idArr []string) (err error) {
			var ids = interfaces(idArr)

			_, err = s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) { // ignore warning

				deleteUserRoleErr := s.connection().WithTx(tx).
					Table("proxy_user_endpoints").
					WhereIn("endpoint_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}
				deleteUserRoleErr = s.connection().WithTx(tx).
					Table("proxy_endpoint_models").
					WhereIn("endpoint_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}
				deleteUserErr := s.connection().WithTx(tx).
					Table("proxy_endpoints").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteUserErr, db.DELETE) {
					return deleteUserErr, nil
				}

				return nil, nil
			})
			return
		})

		info.SetTable("proxy_endpoints").SetTitle("Endpoints").SetDescription("Endpoints")
	}
	formList := userTable.GetForm()

	{
		formList.AddField("Name", "name", db.Varchar, form.Text).FieldMust()
		formList.AddField("Url", "url", db.Varchar, form.Url).FieldMust()
		formList.AddField("Token", "token", db.Varchar, form.Text).FieldMust()
		formList.AddField("Concurrent", "concurrent", db.Int, form.Number).FieldDefault("5")
		formList.AddField("Connections", "connections", db.Int, form.Number).FieldDefault("0")
		// formList.AddField("IsActive", "is_active", db.Bool, form.Switch)
		formList.AddField("RPM", "rpm", db.Int, form.Number).FieldDefault("3")
		formList.AddField("RPD", "rpd", db.Int, form.Number).FieldDefault("200")
		formList.AddField("RequestInMin", "request_in_min", db.Int, form.Number).FieldDefault("0")
		formList.AddField("RequestInDay", "request_in_day", db.Int, form.Number).FieldDefault("0")
		formList.AddField("Models", "models", db.Varchar, form.Select).
			FieldOptionsFromTable("proxy_models", "name", "id").
			FieldDisplay(func(model types.FieldModel) interface{} {
				var endpoints []string

				if model.ID == "" {
					return endpoints
				}
				permissionModel, _ := s.table("proxy_endpoint_models").
					Select("model_id").Where("endpoint_id", "=", model.ID).All()
				for _, v := range permissionModel {
					endpoints = append(endpoints, strconv.FormatInt(v["model_id"].(int64), 10))
				}
				return endpoints
			}).FieldHelpMsg(template.HTML("no corresponding options?") +
			link("/admin/info/endpoints/new", "Create here."))

		formList.AddField("CreatedAt", "created_at", db.Timestamp, form.Datetime).FieldNotAllowEdit().FieldDefault(time.Now().String())
		formList.SetInsertFn(
			func(values form2.Values) (err error) {
				_, err = s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {
					modelId, insertUserRoleErr := s.connection().WithTx(tx).
						Table("proxy_endpoints").
						Insert(dialect.H{
							"name":           values.Get("name"),
							"url":            values.Get("url"),
							"token":          values.Get("token"),
							"concurrent":     values.Get("concurrent"),
							"connections":    values.Get("connections"),
							"rpm":            values.Get("rpm"),
							"rpd":            values.Get("rpd"),
							"request_in_min": values.Get("request_in_min"),
							"request_in_day": values.Get("request_in_day"),

							"created_at": time.Now(),
							"updated_at": time.Now(),
						})

					if db.CheckError(insertUserRoleErr, db.INSERT) {
						return insertUserRoleErr, nil
					}
					subModelsIds := values["models[]"]
					logrus.Info(subModelsIds)
					for _, v := range subModelsIds {
						_, insertUserRoleErr = s.connection().WithTx(tx).
							Table("proxy_endpoint_models").
							Insert(dialect.H{
								"endpoint_id": modelId,
								"model_id":    v,
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
						Table("proxy_endpoints").
						Where("id", "=", modelId).
						Update(dialect.H{
							"name":           values.Get("name"),
							"url":            values.Get("url"),
							"token":          values.Get("token"),
							"concurrent":     values.Get("concurrent"),
							"connections":    values.Get("connections"),
							"rpm":            values.Get("rpm"),
							"rpd":            values.Get("rpd"),
							"request_in_min": values.Get("request_in_min"),
							"request_in_day": values.Get("request_in_day"),

							"updated_at": time.Now(),
						})

					if db.CheckError(insertUserRoleErr, db.UPDATE) {
						return errors.New("update: " + insertUserRoleErr.Error()), nil
					}

					_ = s.connection().WithTx(tx).
						Table("proxy_endpoint_models").
						Where("endpoint_id", "=", modelId).
						Delete()

					subModelIds := values["models[]"]
					for _, v := range subModelIds {
						_, insertUserRoleErr = s.connection().WithTx(tx).
							Table("proxy_endpoint_models").
							Insert(dialect.H{
								"endpoint_id": modelId,
								"model_id":    v,
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

		formList.SetTable("proxy_endpoints").SetTitle("Endpoints").SetDescription("Endpoints")

		formList.SetPostHook(func(values form2.Values) error {
			fmt.Println("userTable.GetForm().PostHook", values)
			return nil
		})
	}

	userTable.GetForm().SetTabGroups(types.
		NewTabGroups("name", "url", "token", "models", "concurrent", "connections", "is_active", "rpm", "rpd", "request_in_min", "request_in_day")).
		// AddGroup("phone", "role", "request_count", "updated_at")).
		SetTabHeaders("New Endpoint")

	return
}
