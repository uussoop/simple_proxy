package database

import (
	"sync"
	"sync/atomic"

	"github.com/rodrikv/openai_proxy/pkg/cache"
	"gorm.io/gorm"
)

type EndpointModelUsage struct {
	gorm.Model
	UserID     uint     `json:"user_id"`
	User       User     `gorm:"foreignKey:UserID"`
	EndpointID uint     `json:"endpoint_id"`
	Endpoint   Endpoint `gorm:"foreignKey:EndpointID"`
	ModelID    uint     `json:"model_id"`
	LLMModel   Model    `gorm:"foreignKey:ModelID"`
	TokenUsed  uint     `json:"token_used"`
}

func (mu *EndpointModelUsage) UniqueIndex() string {
	return "idx_user_endpoint_model"
}

func (mu *EndpointModelUsage) GetOrCreate(u User, e Endpoint, m Model) (created bool, err error) {
	Db.Where("user_id = ? AND endpoint_id = ? AND model_id = ?", u.ID, e.ID, m.ID).First(&mu)

	if mu.ID != 0 {
		return
	}

	mu.UserID = u.ID
	mu.EndpointID = e.ID
	mu.ModelID = m.ID
	mu.TokenUsed = 0

	mu.LLMModel = m
	mu.User = u
	mu.Endpoint = e

	err = Db.Create(&mu).Error

	if err != nil {
		return
	}

	created = true
	return
}

var usageLock sync.Mutex

func (mu *EndpointModelUsage) Increase(tokenCount uint) (err error) {
	usageLock.Lock()
	defer usageLock.Unlock()
	c, ok := cache.GetCache().Get(mu.User.Name + mu.Endpoint.Name + mu.LLMModel.Name)

	var usage uint64

	if ok {
		usage = uint64(c.(uint))
		atomic.AddUint64(&usage, uint64(tokenCount))
	} else {
		Db.First(&mu, mu.ID)
		usage = uint64(mu.TokenUsed)
	}

	err = Db.Model(&mu).Update("token_used", uint(usage)).Error

	cache.GetCache().Set(mu.User.Name+mu.Endpoint.Name+mu.LLMModel.Name, uint(usage), 0)

	return
}
