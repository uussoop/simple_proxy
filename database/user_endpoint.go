package database

import "gorm.io/gorm"

type UserEndpoint struct {
	gorm.Model
	UserID     uint `json:"user_id"`
	EndpointID uint `json:"endpoint_id"`
}

func (ue *UserEndpoint) GetOrCreate() (created bool, err error) {
	err = Db.Where("user_id = ? AND endpoint_id = ?", ue.UserID, ue.EndpointID).First(&ue).Error

	if ue.ID != 0 {
		return false, err
	}

	err = Db.Create(&ue).Error

	return true, err
}

func (ue *UserEndpoint) Update() error {
	result := Db.Save(&ue)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (ue *UserEndpoint) Delete() error {
	result := Db.Delete(&ue)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
