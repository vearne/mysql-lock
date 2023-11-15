package lock

import "time"

type LockCounter struct {
	Name       string    `gorm:"column:name;size:100;index:uni_name,unique" json:"name"`
	Owner      string    `gorm:"column:owner;size:100" json:"owner"`
	Counter    uint64    `gorm:"column:counter" json:"counter"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime NOT NULL;default:CURRENT_TIMESTAMP" json:"createdAt"`
	ModifiedAt time.Time `gorm:"column:modified_at;type:datetime NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"modifiedAt"`
	ExpiredAt  time.Time `gorm:"column:expired_at" json:"expiredAt"`
	// other
	ID int `gorm:"column:id" json:"id"`
}

func (LockCounter) TableName() string {
	return "_lock_counter"
}
