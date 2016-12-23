package service

import (
	"encoding/base64"

	"github.com/fasthttp-contrib/websocket"

	"adbutil"
	"minicap"
)

type MinicapService struct {
	ForwardMinicapPort int
	wsClients          []*websocket.Conn
	SocketService      *minicap.Minicap
	ImgTransferService *ImgTransferService
	WsConnChan         chan *websocket.Conn
}

func NewMinicapService(port int) *MinicapService {
	m := minicap.NewMinicapWithLocalHost(port)
	t := NewImgTransferService(JPG2JPG)
	connChan := make(chan *websocket.Conn, 1)
	clients := make([]*websocket.Conn, 0)
	return &MinicapService{ForwardMinicapPort: port, ImgTransferService: t,
		SocketService: m, WsConnChan: connChan, wsClients: clients}
}

func (m *MinicapService) Start() {
	adbutil.Logger.Debug("start minicap service of %d", m.ForwardMinicapPort)
	var jpgData []byte
	var webpData []byte
	go m.SocketService.Run()
	go m.ImgTransferService.Transfer()
	var wsConn *websocket.Conn
	adbutil.Logger.Debug("MinicapService started...")
	// var wrongIndexs = make([]int, 0)
	for {
		// wrongIndexs = wrongIndexs[:0]
		select {
		case wsConn = <-m.WsConnChan:
			adbutil.Logger.Debug("add new ws client in wsclients")
			m.wsClients = append(m.wsClients, wsConn)
		case jpgData = <-m.SocketService.ImageChan:
			adbutil.Logger.Debug("got jpg data")
			m.ImgTransferService.TData <- jpgData
		case webpData = <-m.ImgTransferService.OutData:
			adbutil.Logger.Debug("==========================================")
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
					// wrongIndexs = append(wrongIndexs, index)
				}
			}
			// if len(wrongIndexs) == 1 {
			// 	m.wsClients = append(m.wsClients[:wrongIndexs[0]], m.wsClients[(wrongIndexs[0]+1):]...)
			// } else if len(wrongIndexs) > 1 {
			// 	panic("have not deal with two clients closed sametime")
			// }
		}
	}
}
