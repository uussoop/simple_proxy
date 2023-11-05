package database

import (
	"strings"
	"sync"
	"time"

	"github.com/rodrikv/openai_proxy/pkg/cache"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"unique"`
	Token string

	Endpoints []Endpoint `gorm:"many2many:user_endpoints;"`
	Models    []Model    `gorm:"many2many:user_models;"`

	RequestCount int
	RateLimit    int  `gorm:"default:3"`
	Limited      bool `gorm:"default:false"`
	TokenLimit   int  `gorm:"default:40000"`
	UsageToday   int
	LastSeen     *time.Time
}

func (u *User) AddEnpoint(e Endpoint) (err error) {
	ens, is := cache.GetCache().Get("endpoint_" + u.Name)
	endpoints := ens.([]Endpoint)

	if is {
		cache.GetCache().Set("endpoint_"+u.Name, append(endpoints, e), time.Minute*5)
	} else {
		cache.GetCache().Set("endpoint_"+u.Name, endpoints, time.Minute*5)
	}
	Db.Model(&u).Association("Endpoints").Append(e)
	return
}

func (u *User) GetEndpoints() (endpoints []Endpoint, err error) {
	ens, is := cache.GetCache().Get("endpoint_" + u.Name)

	if is {
		return ens.([]Endpoint), nil
	}

	err = Db.Model(&u).Association("Endpoints").Find(&endpoints)

	cache.GetCache().Set("endpoint_"+u.Name, endpoints, time.Minute*5)

	return
}

var userRequestLock sync.Mutex

func (u *User) Requested() int {
	key := "request_count:" + u.Name
	c := cache.GetCache()

	requestCount := u.GetRequestCount()

	logrus.Info("requestCount: ", requestCount)

	c.Set(key, requestCount+1, time.Minute*1)
	go func() {
		userRequestLock.Lock()
		defer userRequestLock.Unlock()
		Db.Model(&u).Update("request_count", requestCount+1)
	}()

	return requestCount + 1
}

func (u *User) RemoveRequested() {
	key := "request_count:" + u.Name
	c := cache.GetCache()

	requestCount := u.GetRequestCount()

	if requestCount == 0 {
		return
	}

	c.Set(key, requestCount-1, time.Minute*1)

	go func() {
		userRequestLock.Lock()
		defer userRequestLock.Unlock()
		Db.Model(&u).Update("request_count", requestCount-1)
	}()
}

func (u *User) GetRequestCount() int {
	key := "request_count:" + u.Name
	c := cache.GetCache()
	v, is := c.Get(key)

	if is {
		return v.(int)
	}

	Db.First(&u, u.ID)
	c.Set(key, u.RequestCount, time.Minute*1)
	return u.RequestCount
}

func (u *User) SetLastSeen(t time.Time) {

}

func (u *User) ResetRequestCount() {
	key := "request_count:" + u.Name
	c := cache.GetCache()
	userRequestLock.Lock()
	defer userRequestLock.Unlock()

	c.Set(key, 0, time.Minute*1)

	Db.Model(&u).Update("request_count", 0)
}

func (u *User) HasModel(m Model) bool {
	var models []Model

	Db.Model(&u).Association("Models").Find(&models)

	for _, model := range models {

		if model.ID == m.ID || model.HasSubModel(m.Name) {
			return true
		}
	}

	return false
}

func (user *User) IsRateLimited() bool {
	return user.RateLimit <= user.GetRequestCount()
}

func (user *User) IsLimited() bool {
	if user.Limited {
		return true
	} else {
		if user.UsageToday < user.TokenLimit {
			return false
		} else {
			user.Limited = true
			return true
		}
	}
}

func (u *User) Update() error {
	result := Db.Save(&u)
	if result.Error != nil {
		return result.Error
	}

	return nil
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
	v, is := cache.GetCache().Get(token)

	if is {
		return v.([]User), nil
	}

	var users []User
	result := Db.Where("token = ?", token).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	cache.GetCache().Set(token, users, time.Minute*5)

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
	tokenls := strings.Split(*a, " ")
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
