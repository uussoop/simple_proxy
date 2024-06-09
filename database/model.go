package database

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/uussoop/simple_proxy/pkg/cache"
	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"unique"`
	Token        string
	Limited      bool
	UsageToday   int
	SpecialUsage int `gorm:"default:40000"`
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
func InsertUsers(user []User) error {
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

func UpdateUserUsageToday(userid uint, addedUsage int, reset bool) error {

	var user User
	result := Db.Where("id = ?", userid).Find(&user)
	if reset {
		user.UsageToday = 0
	} else {
		user.UsageToday = addedUsage
	}

	result = Db.Model(&user).Update("usage_today", user.UsageToday)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
func Authenticate(a *string) ([]User, bool) {
	tokenls := strings.Split(*a, " ")
	fmt.Println(tokenls)
	if len(tokenls) > 1 {
		users, userserror := GetUserByToken(strings.TrimSpace(tokenls[1]))
		if userserror != nil {
			return nil, false
		}
		if len(users) != 0 {

			return users, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

func IsLimited(user *User) bool {
	used, ok := cache.GetCache().Get(strconv.Itoa(int(user.ID)))

	if ok {
		user.UsageToday = used.(int)
	} else {
		cache.GetCache().Set(strconv.Itoa(int(user.ID)), user.UsageToday, 0)
	}
	if user.UsageToday <= user.SpecialUsage {
		return false
	} else {
		user.Limited = true
		return true
	}

}

func ResetUsageToday() {
	c := cache.GetCache()

	var users []User
	result := Db.Find(&users)
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	for _, user := range users {
		user.UsageToday = 0
		UpdateUserUsageToday(user.ID, 0, true)
		c.Set(strconv.Itoa(int(user.ID)), 0, 0)
	}

}
