package database

import (
	"strings"

	"gorm.io/gorm"
)

type User struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"unique"`
	Token      string
	Limited    bool
	UsageToday int
}

var Db *gorm.DB

type App struct {
	DB *gorm.DB
}

func InsertUser(user User) error {
	result := Db.Create(&user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func GetAllUsers() ([]User, error) {
	var users []User
	result := Db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

func GetUserByToken(token string) ([]User, error) {

	var users []User
	result := Db.Where("token = ?", token).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}
func UpdateUserUsageToday(user User) error {
	result := Db.Model(&user).Update("usage_today", user.UsageToday)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
func Authenticate(a *string) ([]User, bool) {

	users, userserror := GetUserByToken(strings.Split(*a, " ")[1])
	if userserror != nil {
		return nil, false
	}
	if len(users) != 0 {

		return users, true
	} else {
		return nil, false
	}
}
