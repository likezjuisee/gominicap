# go开发真机远控模块

手机中的minicap进程进行修改，不是stf的原生minicap，在每一帧图片头里添加了rotation，为了兼容原生版本，将rotation字节长度改成了0

## 相关引用项目
* minicap [https://github.com/openstf/minicap](https://github.com/openstf/minicap)
* webp [https://github.com/chai2010/webp](https://github.com/chai2010/webp)
* fasthttp [https://github.com/valyala/fasthttp](https://github.com/valyala/fasthttp)
* fasthttprouter [https://github.com/buaazp/fasthttprouter](https://github.com/buaazp/fasthttprouter)
* websocket [https://github.com/fasthttp-contrib/websocket](https://github.com/fasthttp-contrib/websocket)


## minicap && jpeg2webp

### 管道形式 socket recv image | jpeg2webp | webp websocket 2 web
 
![](http://i.imgur.com/caLaHGe.png)

	for {
		select {
		case wsConn = <-m.WsConnChan:
			adbutil.Logger.Debug("add new ws client in wsclients")
			m.wsClients = append(m.wsClients, wsConn)
		case jpgData = <-m.SocketService.ImageChan:
			adbutil.Logger.Debug("got jpg data")
			m.ImgTransferService.TData <- jpgData
		case webpData = <-m.ImgTransferService.OutData:
			adbutil.Logger.Debug("got webp data length %d", len(webpData))
			if len(webpData) == 0 {
				adbutil.Logger.Debug("webp data is empty")
				continue
			}
			adbutil.Logger.Debug("wsclients length %d", len(m.wsClients))
			for _, ws := range m.wsClients {
				err := ws.WriteMessage(websocket.TextMessage, []byte("data:image/jpeg;base64,"+base64.StdEncoding.EncodeToString(webpData)))
				if err != nil {
					adbutil.Logger.Error("send data to browser failed")
					adbutil.Logger.Error(err.Error())
				}
			}
		}
	}


### jpeg2webp
	func (t *ImgTransferService) transferJPG2Webp(jpgData []byte) (webpData []byte) {
		var err error
		img, err := jpeg.Decode(bytes.NewReader(jpgData))
		if err != nil {
			adbutil.Logger.Error("transferJPG2Webp err: %s", err.Error())
			// SaveJPGFile(jpgData)
			// panic(err)
			return
		}
		webpData, err = webp.EncodeRGB(img, lossrate)
		if err != nil {
			adbutil.Logger.Error(err.Error())
			return
		}
		return
	}
## webserver

### 依赖库
* github.com/valyala/fasthttp
* github.com/buaazp/fasthttprouter
* github.com/fasthttp-contrib/websocket
### handler
#### httphandler
	package httphandler
	
	import (
		"github.com/valyala/fasthttp"
	
		"../servertemplate"
	)
	
	func DetailHandler(ctx *fasthttp.RequestCtx) {
		p := &servertemplate.DetailPage{
			CTX: ctx,
		}
		ctx.SetContentType("text/html; charset=utf-8")
		servertemplate.WritePageTemplate(ctx, p)
	}
	
	func IndexHandler(ctx *fasthttp.RequestCtx) {
		p := &servertemplate.IndexPage{
			CTX: ctx,
		}
		ctx.SetContentType("text/html; charset=utf-8")
		servertemplate.WritePageTemplate(ctx, p)
	}

#### websockethandler
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
 

