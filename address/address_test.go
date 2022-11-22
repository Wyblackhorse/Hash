/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package address

import (
	"fmt"
	"testing"
)

func TestAddress_Bytes(t *testing.T) {


	//fmt.Println(HexToAddress("41f1beb2cb0d70e5d53bc23e96a536d1b0b2799a71"))

	d:=HexToAddress("41f1beb2cb0d70e5d53bc23e96a536d1b0b2799a71")
	fmt.Println(d)
}

