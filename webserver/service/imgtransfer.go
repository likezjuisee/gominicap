package service

import (
	"bytes"
	"image/jpeg"
	"os"

	"github.com/chai2010/webp"

	"adbutil"
)

const (
	JPG2WEBP = 1
	JPG2JPG  = 100
)

var lossrate = float32(80.0)

type ImgTransferService struct {
	TType   int
	TData   chan []byte
	OutData chan []byte
}

func NewImgTransferService(ttype int) *ImgTransferService {
	tData := make(chan []byte, 1000)
	outData := make(chan []byte, 1000)
	return &ImgTransferService{TType: ttype, TData: tData, OutData: outData}
}

func (t *ImgTransferService) Transfer() {
	adbutil.Logger.Debug("start transfer...")
	var data []byte
	var webpData []byte
	for {
		data = <-t.TData
		if t.TType == JPG2WEBP {
			webpData = t.transferJPG2Webp(data)
			t.OutData <- webpData
		} else if t.TType == JPG2JPG {
			t.OutData <- data
		}
	}
}

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

func SaveWebpFile(webData []byte) {
	f, _ := os.Create("output.webp")
	defer f.Close()
	f.Write(webData)
}

func SaveJPGFile(jpgData []byte) {
	f, _ := os.Create("output.jpg")
	defer f.Close()
	f.Write(jpgData)
}
