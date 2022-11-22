/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	tele "gopkg.in/telebot.v3"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func JsonWrite(context *gin.Context, status int, result interface{}, msg string) {
	context.JSON(http.StatusOK, gin.H{
		"code":   status,
		"result": result,
		"msg":    msg,
	})
}
func ReturnError101(context *gin.Context, msg string) {
	JsonWrite(context, -101, map[string]interface{}{}, msg)
}

func ReturnOk200(context *gin.Context, msg string, ) {
	JsonWrite(context, 200, map[string]interface{}{}, msg)
}

func ReturnOk200Data(context *gin.Context, result interface{}, msg string, ) {
	JsonWrite(context, 200, result, msg)
}

func ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}

func GetNiuNiuResultForBanker(hash string) []int {
	startLen := len(hash) - 5
	newHash := hash[startLen:]
	var result int       //庄家结果
	var playerResult int //玩家结果
	dataArray := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for k, v := range newHash {
		if k < 3 {
			a, _ := strconv.Atoi(string(v))
			if InArray(dataArray, a) { //
				result = result + a
			} else {
				result = result + 10
			}
		}
		if k > 1 {
			a, _ := strconv.Atoi(string(v))
			if InArray(dataArray, a) { //
				playerResult = playerResult + a
			} else {
				playerResult = playerResult + 10
			}
		}
	}
	return []int{result, playerResult}
}

/**
数组是否存在摸一个元素
*/
func InArray(arr []int, param int) bool {
	for _, v := range arr {
		if param == v {
			return true
		}
	}
	return false
}

func WsResult(str string) bool {
	start := len(str) - 2
	newHash := str[start:]
	if newHash[0] >= 48 && newHash[0] <= 57 { //第一个是 数字
		if newHash[1] >= 48 && newHash[1] <= 57 {
			return true //玩家赢
		}
	}
	if newHash[0] >= 97 && newHash[0] <= 122 { //第一个是 数字
		if newHash[1] >= 97 && newHash[1] <= 122 {
			return true //玩家赢
		}
	}
	return false
}

/***

返回单双结果  true  玩家赢  false  庄家赢
*/
func DsResult(str string, money float64) bool {
	start := len(str) - 3
	newHash := str[start:]
	ll := ""
	moneyStr := strconv.FormatFloat(money, 'f', 0, 64)
	if find := strings.Contains(moneyStr, "."); find {
		strArray := strings.Split(moneyStr, ".")
		ll = strArray[0]
	} else {
		ll = moneyStr
	}
	//获取下注数字
	betNum, _ := strconv.Atoi(string(ll[len(ll)-1:]))
	bet := "单"
	if betNum%2 == 0 {
		bet = "双"
	} else {
		bet = "单"
	}
	//对 newHash 开始遍历  获取 数字
	initNum := 0
	initBool := false
	for _, v := range newHash {
		if v >= 48 && v <= 57 { //说明是数字
			initBool = true
			initNum, _ = strconv.Atoi(string(v))
		}
	}
	resutlBet := "单"
	if !initBool { //说明全是字母
		return false
	} else {
		if initNum%2 == 0 { //双
			resutlBet = "双"
		} else {
			resutlBet = "单"
		}
	}

	//fmt.Println(bet)  //玩家
	//fmt.Println(resutlBet)  //庄家
	if bet == resutlBet { //玩家赢
		return true
	} else {
		return false
	}
}

/**
  单双返回玩家和庄家的投注结果
*/
func DsResultBetData(str string, money float64) (string, string) {
	start := len(str) - 3
	newHash := str[start:]
	ll := ""
	moneyStr := strconv.FormatFloat(money, 'f', 0, 64)
	if find := strings.Contains(moneyStr, "."); find {
		strArray := strings.Split(moneyStr, ".")
		ll = strArray[0]
	} else {
		ll = moneyStr
	}
	//获取下注数字
	betNum, _ := strconv.Atoi(string(ll[len(ll)-1:]))
	bet := "单"
	if betNum%2 == 0 {
		bet = "双"
	} else {
		bet = "单"
	}
	//对 newHash 开始遍历  获取 数字
	initNum := 0
	initBool := false
	for _, v := range newHash {
		if v >= 48 && v <= 57 { //说明是数字
			initBool = true
			initNum, _ = strconv.Atoi(string(v))
		}
	}
	resutlBet := "单"
	if !initBool { //说明全是字母(这种可能几乎不可以)
		return bet + "@" + strconv.Itoa(betNum), resutlBet + "@" + strconv.Itoa(initNum)
	} else {
		if initNum%2 == 0 { //双
			resutlBet = "双"
		} else {
			resutlBet = "单"
		}
	}
	return bet + "@" + strconv.Itoa(betNum), resutlBet + "@" + strconv.Itoa(initNum)
}

/***

返回  百家乐的结果   返回倍数   0倍(庄家赢)   其他倍数(闲赢)
*/

func BjlResult(str string, betType int) float64 {
	LastStr := str[len(str)-5:]
	//庄家点数
	banker := LastStr[:2]
	player := LastStr[3:]

	bankerDs := 0
	playerDs := 0
	for _, v := range banker {
		c, _ := strconv.Atoi(string(v))
		bankerDs = bankerDs + c
		if bankerDs > 10 {
			bankerDs = bankerDs % 10
		}

	}
	for _, v := range player {
		c, _ := strconv.Atoi(string(v))
		playerDs = playerDs + c
		if playerDs > 10 {
			playerDs = playerDs % 10
		}
	}
	//比较庄家和闲家的点数
	if bankerDs == playerDs {
		if betType == 3 {
			//和  并且玩家赢
			return 8
		} else {
			return 0.999 //扣除手续费(平)
		}
	} else if bankerDs > playerDs && betType == 1 {
		//庄家赢
		return 1.95
	} else if bankerDs < playerDs && betType == 2 {
		//闲家赢
		return 1.95
	}

	return 0
}

/**
  返回百家乐 玩家  庄家结果
*/

func BjlBetResultForTelegram(str string) (string, string) {
	LastStr := str[len(str)-5:]
	//庄家点数
	banker := LastStr[:2]
	player := LastStr[3:]

	bankerDs := 0
	playerDs := 0
	for _, v := range banker {
		c, _ := strconv.Atoi(string(v))
		bankerDs = bankerDs + c
		if bankerDs > 10 {
			bankerDs = bankerDs % 10
		}
	}
	for _, v := range player {
		c, _ := strconv.Atoi(string(v))
		playerDs = playerDs + c
		if playerDs > 10 {
			playerDs = playerDs % 10
		}
	}
	return strconv.Itoa(playerDs), strconv.Itoa(bankerDs)
}

func NiuNumToString(num int) string {
	if num == 1 {
		return "牛一"
	} else if num == 2 {
		return "牛二"
	} else if num == 3 {
		return "牛三"
	} else if num == 4 {
		return "牛四"
	} else if num == 5 {
		return "牛五"
	} else if num == 6 {
		return "牛六"
	} else if num == 7 {
		return "牛七"
	} else if num == 8 {
		return "牛八"
	} else if num == 9 {
		return "牛九"
	} else if num == 10 {
		return "牛十"
	}
	return ""
}

//golang 定时器，启动的时候执行一次，以后每天晚上12点执行
func StartTimer(f func()) {
	go func() {
		for {
			f()
			now := time.Now()
			// 计算下一个零点
			next := now.Add(time.Hour * 24)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())

			fmt.Println(next.Sub(now))
			t := time.NewTimer(next.Sub(now))
			<-t.C
		}
	}()
}

//生成走势图    (结果,背景图地址,)


//调用os.MkdirAll递归创建文件夹
func CreateMuiDir(filePath string) error {
	if !isExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			fmt.Println("创建文件夹失败,error info:", err)
			return err
		}
		return err
	}
	return nil
}

// 判断所给路径文件/文件夹是否存在(返回true是存在)
func isExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

//实时发送给玩家投注结果
func RealTimeSendResult(backgroundUrl string, playerUrl string, bankerUrl string, filePath string, kinds string) bool {
	//获取背景图片
	ImgB, _ := os.Open(backgroundUrl)
	img, _ := jpeg.Decode(ImgB)
	defer ImgB.Close()
	//玩家
	wmb, _ := os.Open(playerUrl)
	watermark, _ := png.Decode(wmb)
	defer wmb.Close()
	//坐标偏差，x轴y轴，这个自己算一下
	offset := image.Pt(46, 546)
	//庄家爱
	three, _ := os.Open(bankerUrl)
	threeWeb, _ := png.Decode(three)
	defer three.Close()

	b := img.Bounds()
	m := image.NewRGBA(b)
	draw.Draw(m, b, img, image.ZP, draw.Src)
	draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)
	draw.Draw(m, threeWeb.Bounds().Add(image.Pt(46, 723)), threeWeb, image.ZP, draw.Over)
	//判断文件夹是否存在不存
	err11 := CreateMuiDir(filePath)
	if err11 != nil {
		return false
	}
	imgW, err1 := os.Create(filePath + "/" + kinds)
	if err1 != nil {
		fmt.Println(err1.Error())
		return false
	}
	err := jpeg.Encode(imgW, m, &jpeg.Options{Quality: jpeg.DefaultQuality})
	if err != nil {
		return false
	}
	defer imgW.Close()
	return true
}

//返回牛牛的图片地址
func NiuReturnImageUrl(v string, kinds int) string {
	//玩家
	if kinds == 1 {
		if v == "牛一" {
			return "picture/Niu/player/1.png"
		} else if v == "牛二" {
			return "picture/Niu/player/2.png"

		} else if v == "牛三" {
			return "picture/Niu/player/3.png"

		} else if v == "牛四" {
			return "picture/Niu/player/4.png"

		} else if v == "牛五" {
			return "picture/Niu/player/5.png"

		} else if v == "牛六" {
			return "picture/Niu/player/6.png"

		} else if v == "牛七" {
			return "picture/Niu/player/7.png"

		} else if v == "牛八" {
			return "picture/Niu/player/8.png"

		} else if v == "牛九" {
			return "picture/Niu/player/9.png"

		} else if v == "牛牛" {
			return "picture/Niu/player/10.png"

		}
	}
	//庄家
	if kinds == 2 {
		if v == "牛一" {
			return "picture/Niu/banker/1.png"
		} else if v == "牛二" {
			return "picture/Niu/banker/2.png"

		} else if v == "牛三" {
			return "picture/Niu/banker/3.png"

		} else if v == "牛四" {
			return "picture/Niu/banker/4.png"

		} else if v == "牛五" {
			return "picture/Niu/banker/5.png"

		} else if v == "牛六" {
			return "picture/Niu/banker/6.png"

		} else if v == "牛七" {
			return "picture/Niu/banker/7.png"

		} else if v == "牛八" {
			return "picture/Niu/banker/8.png"

		} else if v == "牛九" {
			return "picture/Niu/banker/9.png"

		} else if v == "牛牛" {
			return "picture/Niu/banker/10.png"

		}
	}

	return ""
}

//返回单双的图片地址

func DsReturnImageUrl(v string) string {
	ds := strings.Split(v, "@")
	if ds[0] == "单" {
		//单
		return "picture/singleOrDouble/dan/" + ds[1] + ".png"
	} else {
		return "picture/singleOrDouble/both/" + ds[1] + ".png"
	}
}

//发送结果给玩家
func SendPhotoTo(Token string, tId int64, filepath string, win int, gameType string, money string, hash string, backMoney string, created string) {
	pref := tele.Settings{
		Token:  Token, //hash 机器人地址
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := tele.NewBot(pref)
	if err != nil {
		return
	}
	a := tele.User{ID: tId}
	p := &tele.Photo{File: tele.FromDisk(filepath)}
	if win == 2 { //////状态 1输  2 赢 3无效  6 和
		p.Caption = "恭喜您,中奖啦!\n\n游戏类型: " + gameType + "\n游戏金额:" + money + "\n区块哈希:" + "...." + hash[len(hash)-5:] + "\n游戏结果: 赢 \n奖金金额:" + backMoney + "\n时间:" + created
	} else if win == 1 {
		p.Caption = "😭😭真遗憾,您没有中奖\n\n游戏类型: " + gameType + "\n游戏金额:" + money + "\n区块哈希:" + "...." + hash[len(hash)-5:] + "\n游戏结果: 输 \n时间:" + created
	} else if win == 6 {
		p.Caption = "和 \n\n游戏类型: " + gameType + "\n游戏金额:" + money + "\n区块哈希:" + "...." + hash[len(hash)-5:] + "\n游戏结果: 和 \n返还金额:" + backMoney + "\n时间:" + created
	}
	selector := &tele.ReplyMarkup{}
	// Reply buttons.
	btnPrev := selector.URL("验证奖金哈希", "https://tronscan.org/#/transaction/"+hash)
	selector.Inline(
		selector.Row(btnPrev),
	)

	b.Send(&a, p, selector)
}
