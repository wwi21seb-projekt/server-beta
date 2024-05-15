package models

type PushSubscription struct {
	Id       string `gorm:"column:id;primary_key"`
	Username string `gorm:"column:username_fk;type:varchar(20)"`
	User     User   `gorm:"foreignKey:username_fk;references:username"`
	Type     string `gorm:"column:type;type:varchar(20)"` // either "web" or "expo"

}

type VapidKeyResponseDTO struct {
	Key string `json:"key"`
}

type SubscriptionObject struct {
}

type PushSubscriptionRequestDTO struct {
	Type               string             `json:"type" binding:"required"`
	SubscriptionObject SubscriptionObject `json:"subscriptionObject" binding:"required"`
}
