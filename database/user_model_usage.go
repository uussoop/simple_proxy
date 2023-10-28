package database

import (
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

func (mu *EndpointModelUsage) Increase(tokenCount uint) (err error) {
	c, ok := cache.GetCache().Get(mu.User.Name + mu.Endpoint.Name + mu.LLMModel.Name)

	var usage uint

	if ok {
		usage = c.(uint) + tokenCount
	} else {
		Db.First(&mu, mu.ID)
		usage = mu.TokenUsed
	}

	go func() {
		err = Db.Model(&mu).Update("token_used", usage).Error
	}()

	cache.GetCache().Set(mu.User.Name+mu.Endpoint.Name+mu.LLMModel.Name, usage, 0)

	return
}
