/**
 * @Author $
 * @Description //TODO $

 * @Date $ $
 * @Param $
 * @return $
 **/
package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type CommissionModel struct {
	ID          uint   `gorm:"primaryKey;comment:'主键'"`
	FromAddress string //转账的地址
	ToAddress   string //玩家地址(手账地址)
	Created     int64  //我数据库创建的时间
	SuccessTime int64  //成功的区块时间
	Money    float64  `gorm:"type:decimal(10,2)"`  //金额
	CreatedString string   `gorm:"-"`
	SuccessTimeString string   `gorm:"-"`

}

func CheckIsExistModelCommissionModel(db *gorm.DB) {
	if db.HasTable(&CommissionModel{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&CommissionModel{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&CommissionModel{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}
