package database

import (
	"strings"
	"sync"
	"time"

	"github.com/rodrikv/openai_proxy/pkg/cache"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Endpoint struct {
	gorm.Model
	Name         string `json:"name"`
	Url          string `json:"url"`
	Token        string `json:"token"`
	Concurrent   int    `json:"concurrent" gorm:"default:4"`
	Connections  int    `json:"connections"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`
	RPM          int    `json:"rpm" gorm:"type:int;default:3"`
	RPD          int    `json:"rpd" gorm:"type:int;default:200"`
	RequestInMin int    `json:"request_in_min" gorm:"type:int;default:0"`
	RequestInDay int    `json:"request_in_day" gorm:"type:int;default:0"`

	Models []Model `gorm:"many2many:endpoint_models;"`
	Users  []User  `gorm:"many2many:user_endpoints;"`
}

func (e *Endpoint) GetByName(name string) (err error) {
	Db.Where("name = ?", name).First(&e)
	return
}

func (e *Endpoint) Create(name, url, token string, concurrent int, isActive bool, RPM, RPD *int, models *[]Model) (err error) {
	Db.Where("url = ? AND token = ?", url, token).First(&e)

	if e.ID != 0 {
		return
	}

	e.Name = name
	e.Url = url
	e.Token = token
	e.Concurrent = concurrent
	e.IsActive = isActive

	if RPM != nil {
		e.RPM = *RPM
	}

	if RPD != nil {
		e.RPD = *RPD
	}

	if models != nil {
		e.Models = *models
	}

	err = Db.Create(&e).Error

	return
}

func (e *Endpoint) AddUser(user *User) (err error) {
	us := UserEndpoint{
		UserID:     user.ID,
		EndpointID: e.ID,
	}

	created, err := us.GetOrCreate()

	if err != nil {
		return
	}

	if created {
		Db.Model(&e).Association("Users").Append(&user)
	}

	return e.Update()
}

func (e *Endpoint) GetUsers() (users []User, err error) {
	err = Db.Model(&e).Association("Users").Find(&users)

	return
}

var connLock sync.Mutex

func (e *Endpoint) GetConnection() int {
	c := cache.GetCache()
	key := "endpoint:connection:" + e.Name

	v, is := c.Get(key)

	if is {
		return v.(int)
	}

	connLock.Lock()
	defer connLock.Unlock()

	Db.First(&e, e.ID)

	c.Set(key, e.Connections, time.Minute*1)
	return e.Connections
}

func (e *Endpoint) AddConnection() (err error) {
	c := cache.GetCache()
	key := "endpoint:connection:" + e.Name

	conn := e.GetConnection()

	c.Set(key, conn+1, time.Minute*1)
	go func() {
		connLock.Lock()
		defer connLock.Unlock()
		Db.Model(&e).Update("connections", conn+1)
	}()
	return
}

func (e *Endpoint) RemoveConnection() (err error) {
	c := cache.GetCache()
	key := "endpoint:connection:" + e.Name

	conn := e.GetConnection()

	if conn == 0 {
		return
	}

	c.Set(key, conn-1, time.Minute*1)
	go func() {
		connLock.Lock()
		defer connLock.Unlock()
		Db.Model(&e).Update("connections", conn-1)
	}()
	return
}

var requestLock sync.Mutex

func (e *Endpoint) GetRequestInMin() (int, bool) {
	c := cache.GetCache()
	key := "endpoint:request_in_min:" + e.Name

	v, is := c.Get(key)

	if is {
		return v.(int), is
	}
	requestLock.Lock()
	defer requestLock.Unlock()

	Db.First(&e, e.ID)

	return e.RequestInMin, is
}

func (e *Endpoint) GetRequestInDay() (int, bool) {
	c := cache.GetCache()
	key := "endpoint:request_in_day:" + e.Name

	v, is := c.Get(key)

	if is {
		return v.(int), is
	}
	requestLock.Lock()
	defer requestLock.Unlock()

	Db.First(&e, e.ID)

	return e.RequestInDay, is
}

func update_field(e *Endpoint, field string, v int) {
	requestLock.Lock()
	defer requestLock.Unlock()
	Db.Model(&e).Update(field, v)
}

func (e *Endpoint) Requested() {
	c := cache.GetCache()

	requestInMinKey := "endpoint:request_in_min:" + e.Name
	requestInDayKey := "endpoint:request_in_day:" + e.Name

	v1, _ := e.GetRequestInMin()
	logrus.Info("request in min count: ", v1)
	c.Set(requestInMinKey, v1+1, time.Minute*1)

	go update_field(e, "request_in_min", v1+1)

	v2, _ := e.GetRequestInDay()
	logrus.Info("request in day count: ", v2)
	c.Set(requestInDayKey, v2+1, time.Hour*24)

	go update_field(e, "request_in_day", v2+1)
}

func (e *Endpoint) Update() (err error) {
	tx := Db.Save(&e)

	if tx.Error != nil {
		err = tx.Error
	}

	return
}

func (e *Endpoint) DeActivate() {
	cache.GetCache().Set(e.String(), false, time.Minute)
}

func (e *Endpoint) Active() bool {
	v, exists := cache.GetCache().Get(e.String())

	if !exists {
		return e.IsActive
	}

	return v.(bool) && e.IsActive
}

func (e *Endpoint) HasModel(m Model) bool {
	Db.Model(&e).Association("Models").Find(&e.Models)
	logrus.Info(e.Models)
	for _, model := range e.Models {
		if model.ID == m.ID || model.HasSubModel(m.Name) {
			return true
		}
	}

	return false
}

func (e *Endpoint) String() string {
	return strings.ReplaceAll(e.Name+e.Url+e.Token, " ", "_")
}
