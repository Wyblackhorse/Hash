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

//系统配置
type ConfigModel struct {
	ID                  uint   `gorm:"primaryKey;comment:'主键'"`
	SingleOrDoubleHash  string //单双hash
	LuckHash            string //幸运hash
	NiuNiuHash          string //牛牛hash
	BaccaratHash        string //百家乐hash
	AddressOfTheService string //客服地址
}

func CheckIsExistModelConfigModel(db *gorm.DB) {
	if db.HasTable(&ConfigModel{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&ConfigModel{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&ConfigModel{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
		//初始化数据
		db.Save(&ConfigModel{ID: 1})
	}
}
