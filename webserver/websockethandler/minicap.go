package websockethandler

import (
	"fmt"

	"github.com/fasthttp-contrib/websocket"
	"github.com/valyala/fasthttp"

	"../service"
	"adbutil"
)

var MinicapPort = 4102
var minicapService *service.MinicapService

func Minicap(c *websocket.Conn) {
	defer c.Close()
	adbutil.Logger.Debug("new ws client...")
	if minicapService == nil {
		minicapService = service.NewMinicapService(MinicapPort)
		minicapService.WsConnChan <- c
		minicapService.Start()
	} else {
		adbutil.Logger.Debug("add client in list...")
		minicapService.WsConnChan <- c
		adbutil.Logger.Debug("add client in list success")
	}

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if mt == websocket.TextMessage {
			fmt.Println("receive text message...")
			fmt.Println(string(message))
		}
	}
}

var MinicapUpgrader = websocket.New(Minicap)

func MinicapHandler(ctx *fasthttp.RequestCtx) {
	err := MinicapUpgrader.Upgrade(ctx)
	if err != nil {
		adbutil.Logger.Error("upgrade: %s", err.Error())
		return
	}
}
