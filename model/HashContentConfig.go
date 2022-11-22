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

type HashContentConfig struct {
	ID             uint   `gorm:"primaryKey;comment:'主键'"`
	NiuNiu         string `gorm:"type:text"` //牛牛玩法说明
	Baccarat       string `gorm:"type:text"` //百家乐
	SingleOrDouble string `gorm:"type:text"` //单双
	Lucky          string `gorm:"type:text"` //幸运
}

func CheckIsExistModelHashContentConfig(db *gorm.DB) {
	if db.HasTable(&HashContentConfig{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&HashContentConfig{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&HashContentConfig{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
			//数据库的初始化
			h := HashContentConfig{ID: 1}

			h.Add(db)

		}
	}
}

func (h *HashContentConfig) Add(db *gorm.DB) {
	db.Save(&h)
}

func (h *HashContentConfig) GetContent(db *gorm.DB, kinds int) string {
	err := db.Where("id=?", 1).First(&h).Error
	if err != nil {
		return ""
	}

	if kinds == 1 {
		//返回牛牛
		return h.NiuNiu
	}

	if kinds == 2 {

		return h.Baccarat
	}

	if kinds == 3 {

		return h.Lucky
	}

	if kinds == 4 {
		return h.SingleOrDouble
	}

	return ""

}
