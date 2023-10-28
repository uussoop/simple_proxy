package database

import "gorm.io/gorm"

type EndpointModelUsage struct {
	gorm.Model
	UserID     uint `json:"user_id"`
	EndpointID uint `json:"endpoint_id"`
	ModelID    uint `json:"model_id"`
	TokenUsed  uint `json:"token_used"`
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

	err = Db.Create(&mu).Error

	if err != nil {
		return
	}

	created = true
	return
}

func (mu *EndpointModelUsage) Increase(tokenCount uint) (err error) {
	mu.TokenUsed += tokenCount

	err = Db.Save(&mu).Error
	return
}
