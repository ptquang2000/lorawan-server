package models

import (
	"time"
	"fmt"
	"crypto/sha256"

	"gorm.io/gorm"
)

type Gateways struct {
	Id uint32 `gorm:"not null;autoIncrement"`
	Username string `gorm:"type:varchar(100);default:null;uniqueIndex:unique_username"`
	Password_harsh string `gorm:"type:varchar(100);default:null" json:"Password"`
	Salt string `gorm:"type:varchar(35);default:null"`
	Is_superuser bool `gorm:"default:0"`
	Created time.Time `gorm:"type:datetime;default:null"`
}

func ReadGateways() (gateways []Gateways) {
	db.Find(&gateways)
	return
}

func FindGatewayById(id int) (gateway Gateways){
	db.First(&gateway, id)
	return
}

func DeleteGatewayById(id uint32) (tx *gorm.DB) {
	tx = db.Delete(&Gateways{}, id)
	return
}

func (gateway Gateways) Create() (tx *gorm.DB) {
	hash := sha256.New()
	hash.Write([]byte(gateway.Password_harsh + gateway.Salt))
	password_harsh := fmt.Sprintf("%x", hash.Sum(nil))

	gateway.Password_harsh = password_harsh
	gateway.Created = time.Now()

	tx = db.Create(&gateway)
	return
}