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

func (s *SystemTable) GetModelTable(ctx *context.Context) (modelTable table.Table) {

	modelTable = table.NewDefaultTable(table.Config{
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

	info := modelTable.GetInfo().SetFilterFormLayout(form.LayoutThreeCol)
	{
		info.AddField("ID", "id", db.Int).FieldSortable()
		info.AddField("Name", "name", db.Varchar).FieldEditAble(editType.Text).
			FieldFilterable(types.FilterType{Operator: types.FilterOperatorLike})
		info.AddField("Parent", "model_id", db.Varchar).
			FieldJoin(types.Join{
				Table:     "proxy_model_submodels",
				JoinField: "sub_model_id",
				Field:     "id",
			}).
			FieldDisplay(func(model types.FieldModel) interface{} {
				labels := template.HTML("")
				labelTpl := label().SetType("success")

				labelValues := strings.Split(model.Value, types.JoinFieldValueDelimiter)
				for key, label := range labelValues {
					if key == len(labelValues)-1 {
						labels += labelTpl.SetContent(template.HTML(label)).GetContent()
					} else {
						labels += labelTpl.SetContent(template.HTML(label)).GetContent() + " "
					}
				}

				if labels == template.HTML("") {
					return "no roles"
				}

				return labels
			}).FieldFilterable()
		info.AddField("CreatedAt", "created_at", db.Timestamp).
			FieldFilterable(types.FilterType{FormType: form.DatetimeRange})
		info.AddField("UpdatedAt", "updated_at", db.Timestamp).FieldEditAble(editType.Datetime)
		info.AddField("DeletedAt", "deleted_at", db.Timestamp).
			FieldFilterable(types.FilterType{FormType: form.DatetimeRange})

		info.SetTable("proxy_models").SetTitle("Models").SetDescription("Models")

		info.SetDeleteFn(func(idArr []string) (err error) {
			var ids = interfaces(idArr)

			_, err = s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteUserRoleErr := s.connection().WithTx(tx).
					Table("proxy_model_submodels").
					WhereIn("sub_model_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}
				deleteUserRoleErr = s.connection().WithTx(tx).
					Table("proxy_model_submodels").
					WhereIn("model_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}
				deleteUserErr := s.connection().WithTx(tx).
					Table("proxy_models").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteUserErr, db.DELETE) {
					return deleteUserErr, nil
				}

				return nil, nil
			})
			return
		})
	}

	formList := modelTable.GetForm()
	{
		formList.AddField("Name", "name", db.Varchar, form.Text).FieldMust()
		formList.AddField("SubModels", "sub_models", db.Varchar, form.Select).
			FieldOptionsFromTable("proxy_models", "name", "id").
			FieldDisplay(func(model types.FieldModel) interface{} {
				var endpoints []string

				if model.ID == "" {
					return endpoints
				}
				permissionModel, _ := s.table("proxy_model_submodels").
					Select("sub_model_id").Where("model_id", "=", model.ID).All()
				for _, v := range permissionModel {
					endpoints = append(endpoints, strconv.FormatInt(v["sub_model_id"].(int64), 10))
				}
				return endpoints
			}).FieldHelpMsg(template.HTML("no corresponding options?") +
			link("/admin/info/models/new", "Create here."))

		formList.AddField("CreatedAt", "created_at", db.Timestamp, form.Datetime).FieldNotAllowEdit()
		formList.SetTable("proxy_models").SetTitle("Models").SetDescription("Models")

		formList.SetInsertFn(
			func(values form2.Values) (err error) {
				_, err = s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {
					modelId, insertUserRoleErr := s.connection().WithTx(tx).
						Table("proxy_models").
						Insert(dialect.H{
							"name":       values.Get("name"),
							"created_at": time.Now().Format("2006-01-02 15:04:05"),
							"updated_at": time.Now().Format("2006-01-02 15:04:05"),
						})

					if db.CheckError(insertUserRoleErr, db.INSERT) {
						return insertUserRoleErr, nil
					}
					subModelsIds := values["sub_models[]"]
					logrus.Info(subModelsIds)
					for _, v := range subModelsIds {
						_, insertUserRoleErr = s.connection().WithTx(tx).
							Table("proxy_model_submodels").
							Insert(dialect.H{
								"model_id":     modelId,
								"sub_model_id": v,
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
						Table("proxy_models").
						Where("id", "=", modelId).
						Update(dialect.H{
							"name": values.Get("name"),
						})

					if db.CheckError(insertUserRoleErr, db.UPDATE) {
						return errors.New("update: " + insertUserRoleErr.Error()), nil
					}

					_ = s.connection().WithTx(tx).
						Table("proxy_model_submodels").
						Where("model_id", "=", modelId).
						Delete()

					subModelIds := values["sub_models[]"]
					for _, v := range subModelIds {
						_, insertUserRoleErr = s.connection().WithTx(tx).
							Table("proxy_model_submodels").
							Insert(dialect.H{
								"model_id":     modelId,
								"sub_model_id": v,
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

	modelTable.GetForm().SetTabGroups(types.
		NewTabGroups("name", "models", "sub_models", "created_at")).
		// AddGroup("phone", "role", "request_count", "updated_at")).
		SetTabHeaders("New Model")

	return
}
