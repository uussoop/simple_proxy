package database

import (
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	Name      string  `json:"name"`
	SubModels []Model `json:"sub_models" gorm:"many2many:model_submodels;"`
}

func (m *Model) Create(name string, subModels []Model) (err error) {
	Db.Where("name = ?", name).First(&m)

	if m.ID != 0 {
		return
	}

	m.Name = name
	m.SubModels = subModels

	err = Db.Create(&m).Error

	return
}

func (m *Model) Get(name string) (err error) {
	err = Db.Where("name = ?", name).First(&m).Error

	return
}

func (m *Model) HasSubModel(name string) bool {
	Db.Model(&m).Association("SubModels").Find(&m.SubModels)
	for _, sm := range m.SubModels {
		if sm.Name == name || sm.HasSubModel(name) {
			return true
		}
	}

	return false
}
