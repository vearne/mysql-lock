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

type LockCounter struct {
	Name       string    `gorm:"column:name;size:100;index:uni_name,unique" json:"name"`
	Owner      string    `gorm:"column:owner;size:100" json:"owner"`
	Counter    uint64    `gorm:"column:counter" json:"counter"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	ModifiedAt time.Time `gorm:"column:modified_at" json:"modifiedAt"`
	// other
	ID int `gorm:"column:id" json:"id"`
}

func (LockCounter) TableName() string {
	return "_lock_counter"
}
