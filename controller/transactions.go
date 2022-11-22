/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package controller

import "C"
import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

func Transactions(c *gin.Context) {
	address := c.Query("address")
	//url := "https://api.shasta.trongrid.io/v1/accounts/" + address + "/transactions"
	url := "https://api.trongrid.io/v1/accounts/" + address + "/transactions/trc20?only_to=true"
	//https://api.trongrid.io
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(res)
	//fmt.Println(string(body))
	//util.JsonWrite(c, 200, string(body), "获取成功")
	c.String(http.StatusOK, string(body))

}

func GetTransactionInfoById(c *gin.Context) {

	url := "https://api.trongrid.io/wallet/getblockbynum"

	payload := strings.NewReader("{\"num\":39851081}")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	c.String(http.StatusOK, string(body))

}
