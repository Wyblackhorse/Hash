/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package process

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/wangyi/MgHash/dao/mysql"
	"github.com/wangyi/MgHash/dao/redis"
	"github.com/wangyi/MgHash/model"
	"github.com/wangyi/MgHash/tools"
	"github.com/wangyi/MgHash/util"
	tele "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	menu          = &tele.ReplyMarkup{ResizeKeyboard: true}
	btnHelp       = menu.Text("牛牛游戏")
	btnSettings1  = menu.Text("单双游戏")
	btnSettings2  = menu.Text("幸运游戏")
	btnSettings10 = menu.Text("百家乐")
	btnSettings3  = menu.Text("牛牛走势")
	btnSettings4  = menu.Text("单双走势")
	btnSettings5  = menu.Text("幸运走势")
	btnSettings11 = menu.Text("百家乐走势")
	btnSettings6  = menu.Text("福利活动")
	btnSettings7  = menu.Text("联系上线")
	btnSettings8  = menu.Text("推广链接")
	btnSettings9  = menu.Text("个人中心")
	btnSettings12 = menu.Text("盈亏流水")
	btnSettings13 = menu.Text("开启中奖通知")
	btnSettings14 = menu.Text("返回首页")

	/**

	  下面是管理的菜单
	*/

	adminSetNiu = menu.Text("牛牛设置")
	adminSetDs  = menu.Text("单双设置")
	adminSetXy  = menu.Text("幸运设置")
	adminSetBjL = menu.Text("百家乐设置")
	adminSetFl  = menu.Text("福利设置")
	adminSetTg  = menu.Text("推广设置")
	adminSetQf  = menu.Text("群发设置")
	adminBtnQf  = menu.Text("群发")
)

func GoRunning(db *gorm.DB) {
	MassTextingInit(db)
	t := make([]model.Telegram, 0)
	err := db.Where("status= ?", 1).Find(&t).Error
	if err == nil {
		for _, v := range t {
			go func() {
				pref := tele.Settings{
					Token:     v.Token,
					Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
					ParseMode: tele.ModeHTML,
				}
				b, err := tele.NewBot(pref)
				if err != nil {
					log.Fatal(err)
					return
				}
				b.Handle("/start", func(c tele.Context) error {
					user := c.Sender()
					start := c.Message().Text
					if len(start) == 16 {
						//说明是有上级的
						p := model.UserModel{TelegramId: user.ID, TelegramUserName: c.Sender().Username}
						p.EnteringTelegramId(mysql.DB, redis.Rdb, start[7:])
					}
					u := model.UserModel{TelegramId: user.ID}
					returnMsg := u.ReplyTelegramStart(mysql.DB, redis.Rdb)
					if returnMsg == "请选择快捷操作!" {
						return c.Send(returnMsg, SetMenu())
					}
					return c.Reply(returnMsg)
				})
				b.Handle(tele.OnText, func(c tele.Context) error {
					user := c.Sender()
					B, _ := redis.Rdb.HExists("TelegramId_"+strconv.Itoa(int(user.ID)), "status").Result()
					if B == false {
						return nil
					}
					status, _ := redis.Rdb.HGet("TelegramId_"+strconv.Itoa(int(user.ID)), "status").Result()
					if status == "1" {
						redis.Rdb.HDel("TelegramId_"+strconv.Itoa(int(user.ID)), "status")
						//入库操作
						if len(c.Text()) > 34 {
							return c.Send("您称不合法,请重新输入")
						}

						P1 := model.UserModel{}
						//判断 这个 飞机 id 是否已经存在了
						err3 := mysql.DB.Where("telegram_id=?", user.ID).First(&P1).Error
						if err3 == nil {

							err3 := mysql.DB.Model(&model.UserModel{}).Where("id=?", P1.ID).Update(&model.UserModel{Nickname: c.Text()}).Error
							if err3 != nil {
								return c.Send("入库失败,请稍后重试")
							}
							redis.Rdb.HSet("TelegramId_"+strconv.Itoa(int(user.ID)), "status", 2) //1 输入昵称
							return c.Send("请输入TRX地址")
						} else {
							uu := model.UserModel{TelegramId: user.ID, Created: time.Now().Unix(), Nickname: c.Text(), TelegramUserName: user.Username}
							err2 := mysql.DB.Save(&uu).Error
							if err2 != nil {
								return c.Send("入库失败,请稍后重试")
							}
							redis.Rdb.HSet("TelegramId_"+strconv.Itoa(int(user.ID)), "status", 2) //1 输入昵称
							return c.Send("请输入TRX地址")
						}
					}
					if status == "2" {
						//输入地址的内容

						if len(c.Text()) != 34 {
							return c.Send("请输入正确的TRX地址")
						}
						//删除
						redis.Rdb.HDel("TelegramId_"+strconv.Itoa(int(user.ID)), "status")
						//判断地址是否存在
						uu := model.UserModel{}
						err2 := mysql.DB.Where("trc20_address=?", c.Text()).First(&uu).Error
						if err2 == nil {
							//存在
							//账号被人绑定了
							if uu.TelegramId != 0 && uu.Nickname != "" {
								redis.Rdb.HSet("TelegramId_"+strconv.Itoa(int(user.ID)), "status", 2) //1 输入昵称
								return c.Send("对不起,改地址已经被别人绑定,请输入其他地址!")
							}
							//账号没有人绑定
							if uu.TelegramId == 0 && uu.Nickname == "" {
								//账号没有人绑定
								oo := model.UserModel{}
								err2 := mysql.DB.Where("telegram_id=?", user.ID).First(&oo).Error
								if err2 == nil {
									update := model.UserModel{
										Nickname:         oo.Nickname,
										TelegramId:       oo.TelegramId,
										TelegramUserName: oo.TelegramUserName,
									}
									//开启事务
									db := mysql.DB.Begin()
									//删除 oo 的值
									err2 := db.Where("id=?", oo.ID).Delete(&model.UserModel{}).Error
									if err2 != nil {
										db.Rollback()
										return c.Send("入库失败,请稍后重试")
									}
									//更新数据
									err2 = db.Model(&model.UserModel{}).Where("id=?", uu.ID).Update(&update).Error
									if err2 != nil {
										db.Rollback()
										return c.Send("入库失败,请稍后重试")
									}

									db.Commit()
									//入库成功
									return c.Send("success", SetMenu())
								}
							}

						} else {
							//不存在   直接更新!
							oo := model.UserModel{}
							err2 := mysql.DB.Where("telegram_id=?", user.ID).First(&oo).Error
							if err2 != nil {
								return c.Send("入库失败,请稍后重试")
							}
							update := model.UserModel{
								Trc20Address:   c.Text(),
								InvitationCode: tools.CheckRandStringRunesIsRepetition(redis.Rdb),
							}
							err2 = mysql.DB.Model(&model.UserModel{}).Where("id=?", oo.ID).Update(update).Error
							if err2 != nil {
								return c.Send("入库失败,请稍后重试")
							}

							redis.Rdb.HSet("CheckRandStringRunesIsRepetition", update.InvitationCode, update.Trc20Address)
							//入库成功
							return c.Send("success", SetMenu())

						}

					}

					return nil
				})
				//牛牛游戏
				b.Handle(&btnHelp, func(c tele.Context) error {
					aa := viper.GetString("hash.NiuNiuCommon")
					//p := &tele.Photo{File: tele.FromDisk("506.jpg")}
					pBool, _ := redis.Rdb.HExists("SystemConfig", "setNnPicture").Result()
					p := &tele.Photo{}
					if pBool {
						idF, _ := redis.Rdb.HGet("SystemConfig", "setNnPicture").Result()
						//p = &tele.Photo{File: tele.FromURL(idF)}
						p.File = tele.FromURL(idF)
					} else {
						p.File = tele.FromDisk("506.jpg")
					}
					p.Caption, _ = redis.Rdb.HGet("SystemConfig", "setNnText").Result()
					p.Caption = strings.Replace(p.Caption, "@@@", aa, 1)
					_ = c.Reply(p)
					_ = c.Send(aa)
					return nil
				})
				//单双游戏
				b.Handle(&btnSettings1, func(c tele.Context) error {
					aa := viper.GetString("hash.DsCommon")
					//p := &tele.Photo{File: tele.FromDisk("508.jpg")}

					pBool, _ := redis.Rdb.HExists("SystemConfig", "setDsPicture").Result()
					p := &tele.Photo{}
					if pBool {
						idF, _ := redis.Rdb.HGet("SystemConfig", "setDsPicture").Result()
						//p = &tele.Photo{File: tele.FromURL(idF)}
						p.File = tele.FromURL(idF)
					} else {
						p.File = tele.FromDisk("508.jpg")
					}
					//p := &tele.Photo{File: tele.FromURL("AgACAgUAAxkBAAIE0GKWS8M0f7of_dS_MEu9KZa82YWiAAI7sDEb_VqxVMH9cY3by-zuAQADAgADeAADJAQ")}
					p.Caption, _ = redis.Rdb.HGet("SystemConfig", "setDsText").Result()
					p.Caption = strings.Replace(p.Caption, "@@@", aa, 1)
					_ = c.Reply(p)
					_ = c.Send(aa)
					return nil
				})
				//双尾游戏(幸运游戏)
				b.Handle(&btnSettings2, func(c tele.Context) error {
					aa := viper.GetString("hash.XyCommon")
					//p := &tele.Photo{File: tele.FromDisk("507.jpg")}

					pBool, _ := redis.Rdb.HExists("SystemConfig", "setXyPicture").Result()
					p := &tele.Photo{}
					if pBool {
						idF, _ := redis.Rdb.HGet("SystemConfig", "setXyPicture").Result()
						//p = &tele.Photo{File: tele.FromURL(idF)}
						p.File = tele.FromURL(idF)
					} else {
						p.File = tele.FromDisk("507.jpg")

					}
					//p := &tele.Photo{File: tele.FromURL("AgACAgUAAxkBAAIE4WKWTAWqnUSvgnYvvkB8M_UtlDdlAAI6sDEb_VqxVD6WVHWZXgoqAQADAgADeAADJAQ")}
					p.Caption, _ = redis.Rdb.HGet("SystemConfig", "setXyText").Result()
					p.Caption = strings.Replace(p.Caption, "@@@", aa, 1)
					_ = c.Reply(p)
					_ = c.Send(aa)
					return nil
				})
				//百家乐 btnSettings10
				b.Handle(&btnSettings10, func(c tele.Context) error {
					aa := viper.GetString("hash.BjLCommon")
					//p := &tele.Photo{File: tele.FromDisk("505.jpg")}
					pBool, _ := redis.Rdb.HExists("SystemConfig", "setBjLPicture").Result()
					p := &tele.Photo{}
					if pBool {
						idF, _ := redis.Rdb.HGet("SystemConfig", "setBjLPicture").Result()
						//p = &tele.Photo{File: tele.FromURL(idF)}
						p.File = tele.FromURL(idF)
					} else {
						p.File = tele.FromDisk("505.jpg")

					}
					//p := &tele.Photo{File: tele.FromURL("AgACAgUAAxkBAAIYI2KWSUGZskEZtlTIqlXUM08tSeOmAAL8sjEbkPiwVL7RPNFWiuDyAQADAgADcwADJAQ")}
					p.Caption, _ = redis.Rdb.HGet("SystemConfig", "setBjLText").Result()
					p.Caption = strings.Replace(p.Caption, "@@@", aa, 1)
					_ = c.Reply(p)
					_ = c.Send(aa)
					return nil
				})
				//牛牛走势  btnSettings3
				b.Handle(&btnSettings3, func(c tele.Context) error {
					//查询地址
					u := model.UserModel{TelegramId: c.Sender().ID}
					address, err := u.ReturnTrxAddress(mysql.DB)
					if err != nil {
						return c.Reply("对不起,系统错误")
					}
					t := model.TransactionRecordModel{Form: address, BetName: "牛牛"}
					tt := t.GetTrend(mysql.DB)
					if len(tt) == 0 {
						return c.Reply("您还没有投注记录")
					}
					filePath := "trend/" + strconv.FormatInt(c.Sender().ID, 10)
					kinds := "NiuTrend.png"
					strBool := util.SetTrendPicture("picture/Niu/back.jpg", tt, filePath, kinds)
					if strBool == false {
						return c.Reply("您还没有投注记录")
					}
					p := &tele.Photo{File: tele.FromDisk(filePath + "/" + kinds)}
					p.Caption = ""
					return c.Reply(p)
				})
				//单双走势  btnSettings4
				b.Handle(&btnSettings4, func(c tele.Context) error {
					//查询地址
					u := model.UserModel{TelegramId: c.Sender().ID}
					address, err := u.ReturnTrxAddress(mysql.DB)
					if err != nil {
						return c.Reply("对不起,系统错误")
					}
					t := model.TransactionRecordModel{Form: address, BetName: "单双"}
					tt := t.GetTrend(mysql.DB)
					if len(tt) == 0 {
						return c.Reply("您还没有投注记录")
					}
					filePath := "trend/" + strconv.FormatInt(c.Sender().ID, 10)
					kinds := "DsTrend.png"
					strBool := util.SetTrendPicture("picture/singleOrDouble/back.jpg", tt, filePath, kinds)
					if strBool == false {
						return c.Reply("您还没有投注记录")
					}
					p := &tele.Photo{File: tele.FromDisk(filePath + "/" + kinds)}
					p.Caption = ""
					return c.Reply(p)
				})
				//双尾走势(幸运)
				b.Handle(&btnSettings5, func(c tele.Context) error {
					//查询地址
					u := model.UserModel{TelegramId: c.Sender().ID}
					address, err := u.ReturnTrxAddress(mysql.DB)
					if err != nil {
						return c.Reply("对不起,系统错误")
					}
					t := model.TransactionRecordModel{Form: address, BetName: "幸运"}
					tt := t.GetTrend(mysql.DB)
					if len(tt) == 0 {
						return c.Reply("您还没有投注记录")
					}
					filePath := "trend/" + strconv.FormatInt(c.Sender().ID, 10)
					kinds := "LuckTrend.png"
					strBool := util.SetTrendPicture("picture/Luckly/back.jpg", tt, filePath, kinds)
					if strBool == false {
						return c.Reply("您还没有投注记录")
					}
					p := &tele.Photo{File: tele.FromDisk(filePath + "/" + kinds)}
					p.Caption = ""
					return c.Reply(p)
				})
				//百家乐走势
				b.Handle(&btnSettings11, func(c tele.Context) error {
					//查询地址
					u := model.UserModel{TelegramId: c.Sender().ID}
					address, err := u.ReturnTrxAddress(mysql.DB)
					if err != nil {
						return c.Reply("对不起,系统错误")
					}
					t := model.TransactionRecordModel{Form: address, BetName: "百家乐"}
					tt := t.GetTrend(mysql.DB)
					if len(tt) == 0 {
						return c.Reply("您还没有投注记录")
					}
					filePath := "trend/" + strconv.FormatInt(c.Sender().ID, 10)
					kinds := "LuckTrend.png"
					strBool := util.SetTrendPicture("picture/Baccarat/back.jpg", tt, filePath, kinds)
					if strBool == false {
						return c.Reply("您还没有投注记录")
					}
					p := &tele.Photo{File: tele.FromDisk(filePath + "/" + kinds)}
					p.Caption = ""
					return c.Reply(p)
				})
				//福利活动
				b.Handle(&btnSettings6, func(c tele.Context) error {
					//aa := viper.GetString("hash.WebUrl")
					//p := &tele.Photo{File: tele.FromDisk("welfare.png")}
					pBool, _ := redis.Rdb.HExists("SystemConfig", "setFlPicture").Result()
					p := &tele.Photo{}
					if pBool {
						idF, _ := redis.Rdb.HGet("SystemConfig", "setFlPicture").Result()
						//p = &tele.Photo{File: tele.FromURL(idF)}
						p.File = tele.FromURL(idF)
					} else {
						p.File = tele.FromDisk("welfare.png")
					}
					//p := &tele.Photo{File: tele.FromURL("AgACAgUAAxkBAAIE5mKWTF4hvJn8v2JblTC7MnjRNNASAAI8sDEb_VqxVP5JWe-TyBJCAQADAgADeQADJAQ")}
					p.Caption, _ = redis.Rdb.HGet("SystemConfig", "setFlText").Result()
					return c.Reply(p)
				})
				//推广链接
				b.Handle(&btnSettings8, func(c tele.Context) error {
					u := model.UserModel{TelegramId: c.Sender().ID}
					code, err := u.ReturnInCode(mysql.DB)
					if err != nil {
						return c.Send("系统错误,请联系管理员")
					}
					//TglJ := "https://t.me/XiangjiaoBot?start=" + code
					//p := &tele.Photo{File: tele.FromDisk("promotion.jpg")}

					pBool, _ := redis.Rdb.HExists("SystemConfig", "setTgPicture").Result()
					p := &tele.Photo{}
					if pBool {
						idF, _ := redis.Rdb.HGet("SystemConfig", "setTgPicture").Result()
						//p = &tele.Photo{File: tele.FromURL(idF)}
						p.File = tele.FromURL(idF)
					} else {
						p.File = tele.FromDisk("promotion.png")
					}
					//p := &tele.Photo{File: tele.FromURL("AgACAgUAAxkBAAIE6GKWTJgudlkTHMOspnDKR54467ZyAAIUrzEbWpawVHSsIdckKMetAQADAgADeQADJAQ")}
					//gw := viper.GetString("hash.WebUrl")
					//pd := viper.GetString("hash.Channel")
					//p.Caption = "您的邀请链接是\n" + TglJ + "\n\n\n\n芒果官网：" + gw + "\n芒果频道：" + pd + ""

					p.Caption, _ = redis.Rdb.HGet("SystemConfig", "setTgText").Result() //@@@@
					p.Caption = strings.Replace(p.Caption, "@@@@", code, 1)
					return c.Reply(p)
				})
				//联系上线
				b.Handle(&btnSettings7, func(c tele.Context) error {
					//不存在   直接更新!
					oo := model.UserModel{}

					err2 := mysql.DB.Where("telegram_id=?", c.Sender().ID).First(&oo).Error
					if err2 != nil {
						return c.Reply("对不起,系统错误")
					}

					if oo.SuperiorId == 0 {
						return c.Reply("您没有代理")
					}

					op := model.UserModel{}
					err2 = mysql.DB.Where("id=?", oo.SuperiorId).First(&op).Error
					if err2 != nil {
						return c.Reply("对不起,系统错误")
					}
					if op.TelegramUserName == "" {
						fmt.Println(op.TelegramUserName)
						return c.Reply("您没有代理")
					}
					url := "https://t.me/" + op.TelegramUserName
					selector := &tele.ReplyMarkup{}
					// Reply buttons.
					btnPrev := selector.URL("祁同伟专线", url)
					selector.Inline(
						selector.Row(btnPrev),
					)
					return c.Reply("点击联系代理", selector)
				})
				//个人中心
				b.Handle(&btnSettings9, func(c tele.Context) error {
					return c.Reply("个人中心", SetMenuTwo())
				})
				//亏盈流水 btnSettings12
				b.Handle(&btnSettings12, func(c tele.Context) error {
					//查询地址
					u := model.UserModel{TelegramId: c.Sender().ID}
					address, err := u.ReturnTrxAddress(mysql.DB)
					if err != nil {
						return c.Reply("对不起,系统错误")
					}
					// 今日 和昨日的 盈利
					tt := model.TransactionRecordModel{Form: address}
					return c.Reply(tt.GetTodayOrYestDayGetMoney(mysql.DB))
				})
				//返回首页
				b.Handle(&btnSettings14, func(c tele.Context) error {
					return c.Reply("返回首页", SetMenu())
				})
				//中奖通知 btnSettings13
				b.Handle(&btnSettings13, func(c tele.Context) error {
					return c.Reply("您已经开启了您你中奖通知")
				})
				b.Handle(tele.OnPhoto, func(c tele.Context) error {
					//判断权限
					u := model.UserModel{TelegramId: c.Sender().ID}
					if u.IfAdminForTid(mysql.DB) == true {
						//获取可以设置的图片
						aa := ReturnPlayKinks()
						if aa == "" {
							return c.Send("请先点击你要设置的选项!")
						}
						photo := c.Message().Photo
						fmt.Println(photo.FileID)
						redis.Rdb.HSet("SystemConfig", "set"+aa+"Picture", photo.FileID)
						redis.Rdb.HSet("SystemConfig", "set"+aa+"Text", c.Text())
						redis.Rdb.HSet("SystemConfigIfUploadPicture", aa, false)
						return c.Send(aa + "游戏设置成功")
					}
					return nil
				})
				//b.Handle(tele.OnText, func(c tele.Context) error {
				//	redis.Rdb.HSet("TextConfig", "status", c.Text())
				//	pp, _ := redis.Rdb.HGet("TextConfig", "status").Result()
				//	return c.Reply(pp)
				//})
				b.Handle(&adminSetNiu, func(c tele.Context) error {
					SetPictureIfUpload("Nn")
					return c.Reply("请上传牛牛游戏图片和内容")
				})

				b.Handle(&adminSetDs, func(c tele.Context) error {
					SetPictureIfUpload("Ds")
					return c.Reply("请上传单双游戏图片和内容")
				})

				b.Handle(&adminSetXy, func(c tele.Context) error {
					SetPictureIfUpload("Xy")
					return c.Reply("请上传幸运游戏图片和内容")
				})
				b.Handle(&adminSetBjL, func(c tele.Context) error {
					SetPictureIfUpload("BjL")
					return c.Reply("请上传百家乐游戏图片和内容")
				})

				b.Handle(&adminSetFl, func(c tele.Context) error {
					SetPictureIfUpload("Fl")
					return c.Reply("请上传福利图片和内容")
				})

				b.Handle(&adminSetTg, func(c tele.Context) error {
					SetPictureIfUpload("Tg")
					return c.Reply("请上传推广图片和内容")
				})

				//群发   adminSetQf
				b.Handle(&adminSetQf, func(c tele.Context) error {
					SetPictureIfUpload("Qf")
					return c.Reply("请设置群发内容")
				})

				b.Handle(&adminBtnQf, func(c tele.Context) error {
					c.Reply("正在准备群发..........")
					b, _ :=tele.NewBot(pref)
					//获取可以群发的飞机
					data, _ := redis.Rdb.HGetAll("MassTexting").Result()
					if len(data) > 0 {
						for i, _ := range data {
							tID, _ := strconv.ParseInt(i, 10, 64)
							fmt.Println(tID)
							a := tele.User{ID: tID}
							idF, _ := redis.Rdb.HGet("SystemConfig", "setQfPicture").Result()
							p := &tele.Photo{File: tele.FromURL(idF)}
							p.Caption, _ = redis.Rdb.HGet("SystemConfig", "setQfText").Result()
							_, _ = b.Send(&a,p)
						}
					}
					return c.Send("群发完毕,一共发送玩家数量:" + strconv.Itoa(len(data)))
				})

				b.Handle("/SystemAdmin", func(c tele.Context) error {
					msg := c.Text()
					if len(msg) > 14 {
						adminAndPassword := msg[13:]
						dataArray := strings.Split(adminAndPassword, "@")
						if len(dataArray) == 2 {
							username := dataArray[0]
							password := dataArray[1]
							u := model.UserModel{TelegramId: c.Sender().ID, UserName: username, Password: password}
							if u.IfAdmin(mysql.DB) == true {
								return c.Reply("密码输入正确", SetMenuForAdmin())
							}
						}
					}
					return nil
				})

				b.Start()
			}()

		}
	}
}

//开启飞机
func StartPlane(b *tele.Bot) {
	b.Start()

}

//群发初始化

func MassTextingInit(db *gorm.DB) {
	t := make([]model.UserModel, 0)
	db.Where("telegram_id  >  ?", 0).Find(&t)
	if len(t) > 0 {
		for _, i2 := range t {
			redis.Rdb.HSet("MassTexting", strconv.FormatInt(i2.TelegramId, 10), "1")
		}
	}
}

/**
  对普通用户
*/
func SetMenu() *tele.ReplyMarkup {
	var a []tele.Btn
	a = append(a, btnHelp)
	a = append(a, btnSettings1)
	a = append(a, btnSettings2)
	a = append(a, btnSettings10)
	a = append(a, btnSettings3)
	a = append(a, btnSettings4)
	a = append(a, btnSettings5)
	a = append(a, btnSettings11)
	a = append(a, btnSettings6)
	a = append(a, btnSettings7)
	a = append(a, btnSettings8)
	a = append(a, btnSettings9)
	menu.Reply(
		menu.Split(4, a)...,
	)
	return menu
}

func SetMenuForAdmin() *tele.ReplyMarkup {
	var a []tele.Btn
	a = append(a, adminSetNiu)
	a = append(a, adminSetDs)
	a = append(a, adminSetXy)
	a = append(a, adminSetBjL)
	a = append(a, adminSetFl)
	a = append(a, adminSetTg)
	a = append(a, adminSetQf)
	a = append(a, adminBtnQf)
	menu.Reply(
		menu.Split(3, a)...,
	)
	return menu
}

func SetMenuTwo() *tele.ReplyMarkup {

	var a []tele.Btn
	a = append(a, btnSettings12)
	a = append(a, btnSettings7)
	a = append(a, btnSettings13)
	a = append(a, btnSettings14)
	menu.Reply(
		menu.Split(3, a)...,
	)
	return menu
}

//设置  某个图片是否可以上传
func SetPictureIfUpload(playKinds string) {
	//redis.Rdb.HSet("SystemConfigIfUploadPicture","NiNiu",true)
	data, _ := redis.Rdb.HGetAll("SystemConfigIfUploadPicture").Result()
	for k, v := range data {
		if v == "1" {
			//真
			redis.Rdb.HSet("SystemConfigIfUploadPicture", k, false)
		}
	}
	redis.Rdb.HSet("SystemConfigIfUploadPicture", playKinds, true)
}

//获取可以设置的玩法类型的(上传图片的判断!)
func ReturnPlayKinks() string {
	data, _ := redis.Rdb.HGetAll("SystemConfigIfUploadPicture").Result()
	for k, v := range data {
		if v == "1" {
			//
			return k
		}
	}
	return ""
}
