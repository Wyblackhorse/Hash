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
	eeor "github.com/wangyi/MgHash/error"
	"time"
)

//电报定义
type Telegram struct {
	ID      uint   `gorm:"primaryKey;comment:'主键'"`
	Token   string //飞机token
	Name    string //飞机名字
	Status  int    //1启用 2 禁用
	Created int64
}

func CheckIsExistModelTelegram(db *gorm.DB) {
	if db.HasTable(&Telegram{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Telegram{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Telegram{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

func (t *Telegram) Add(Db *gorm.DB) (bool, error) {

	err := Db.Where("token=?", t.Token).First(&Telegram{}).Error
	if err == nil {
		return false, eeor.OtherError("不要重复添加")
	}
	t.Created = time.Now().Unix()
	t.Status = 1
	err = Db.Save(&t).Error
	if err != nil {
		return false, err
	}
	return true, nil
}
