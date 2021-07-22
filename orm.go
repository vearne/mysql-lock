package lock

import "time"

type LockStore struct {
	Name      string    `gorm:"column:name;size:100;index:uni_name,unique" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	// other
	ID int `gorm:"column:id" json:"id"`
}

func (LockStore) TableName() string {
	return "_lock_store"
}
