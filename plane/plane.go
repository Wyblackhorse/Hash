/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package plane

import (
	tele "gopkg.in/telebot.v3"
	"log"
	"time"
)

var (
	BigPlane *tele.Bot
	err      error
)

func Init() error {
	pref := tele.Settings{
		Token:     "5445640379:AAETpJt9-ZxrZzfpLp38X4S7t5VHGq-aK0A",
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tele.ModeHTML,
	}
	BigPlane, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return err
	}
	BigPlane.Handle("/start", func(c tele.Context) error {
		return c.Reply("11212")
	})



	return nil
}
