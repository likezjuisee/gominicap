package minicap

import (
	"adbutil"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	MinicapHeaderLength      = 24
	tmpBufferLength          = 4096 * 2
	ImageFrameRotationLength = uint(0) // 本项目在图片frame的头里添加了一个四字节的整形数字，表示图片的旋转角度；兼容stf minicap，这里设置成0
	ImageFrameDataLength     = uint(4)
	ImageFrameHeaderLength   = ImageFrameRotationLength + ImageFrameDataLength
)

type MinicapHeader struct {
	Version     uint8
	Size        uint8
	Pid         uint32
	RDWidth     uint32
	RDHeight    uint32
	VDWidth     uint32
	VDHeight    uint32
	Orientation uint8
	Quirk       uint8
}

func NewMinicapHeader(header []byte) *MinicapHeader {
	if len(header) != MinicapHeaderLength {
		adbutil.Logger.Error("minicap header length is wrong")
		return nil
	}
	var minicapHeader = &MinicapHeader{}
	minicapHeader.Version = uint8(header[0])
	minicapHeader.Size = uint8(header[1])
	minicapHeader.Pid = binary.LittleEndian.Uint32(header[2:6])
	minicapHeader.RDWidth = binary.LittleEndian.Uint32(header[6:10])
	minicapHeader.RDHeight = binary.LittleEndian.Uint32(header[10:14])
	minicapHeader.VDWidth = binary.LittleEndian.Uint32(header[14:18])
	minicapHeader.VDHeight = binary.LittleEndian.Uint32(header[18:22])
	minicapHeader.Orientation = uint8(header[22])
	minicapHeader.Quirk = uint8(header[23])
	return minicapHeader
}

func (mh *MinicapHeader) String() string {
	return fmt.Sprintf("Version:%d Size:%d Pid:%d RDWidth:%d RDHeight:%d VDWidth:%d VDHeight:%d Orientation:%d Quirk:%d",
		mh.Version, mh.Size, mh.Pid, mh.RDWidth, mh.RDHeight, mh.VDWidth, mh.VDHeight, mh.Orientation, mh.Quirk)
}

type Minicap struct {
	Host             string
	Port             int
	conn             net.Conn
	Header           *MinicapHeader
	imageTmpBuffer   []byte
	rotation         int
	imageFrameLength uint32
	ImageChan        chan []byte
	ImageIndex       int
}

func NewMinicap(host string, port int) *Minicap {
	tmpBuffer := make([]byte, tmpBufferLength)
	imageChan := make(chan []byte, 1)
	return &Minicap{Host: host, Port: port, imageTmpBuffer: tmpBuffer, ImageChan: imageChan, ImageIndex: 0}
}

func NewMinicapWithLocalHost(port int) *Minicap {
	return NewMinicap("127.0.0.1", port)
}

func (m *Minicap) connect() bool {
	d := net.Dialer{
		Timeout: time.Second * 30,
	}
	var err error
	m.conn, err = d.Dial("tcp", m.Host+":"+strconv.Itoa(m.Port))
	if err != nil {
		adbutil.Logger.Error("connect to minicap server failed")
		adbutil.Logger.Error(err.Error())
		return false
	}
	adbutil.Logger.Debug("minicap connect success")
	return true
}

func (m *Minicap) disconnect() bool {
	if m.conn == nil {
		return false
	}
	err := m.conn.Close()
	if err != nil {
		adbutil.Logger.Error("disconnect failed")
		adbutil.Logger.Error(err.Error())
		return false
	}
	adbutil.Logger.Debug("minicap disconnect success")
	return true
}

func (m *Minicap) InitMinicapHeader() (rst bool) {
	buffer := make([]byte, MinicapHeaderLength)
	len := 0
	for len < MinicapHeaderLength {
		adbutil.Logger.Debug("InitMinicapHeader read ...")
		readLen, err := m.conn.Read(buffer)
		adbutil.Logger.Debug("readLen %d", readLen)
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				adbutil.Logger.Error("InitMinicapHeader err: %s", err.Error())
				return
			}
		} else {
			len += readLen
			m.Header = NewMinicapHeader(buffer[:readLen])
			adbutil.Logger.Debug(m.Header.String())
		}
	}
	return true
}

func (m *Minicap) getImageFrameRotation(imageData []byte) (rotation int) {
	r := binary.LittleEndian.Uint32(imageData)
	rotation = int(r)
	return
}

func (m *Minicap) getImageDataLength(imageData []byte) (length uint) {
	l := binary.LittleEndian.Uint32(imageData)
	length = uint(l)
	return
}

func (m *Minicap) readSocketDataUntilLength(length uint) (targetData, leftData []byte) {
	if length <= 0 {
		return
	}
	// adbutil.Logger.Debug("read full data length is %d", length)
	cursor := uint(0)
	for {
		__readLen, err := m.conn.Read(m.imageTmpBuffer)
		if err != nil {
			adbutil.Logger.Error("readSocketDataUntilLength err: %s", err.Error())
			return
		}
		readLen := uint(__readLen)
		// adbutil.Logger.Debug("readLen: %d cursor: %d", readLen, cursor)
		if readLen <= length-cursor {
			cursor += readLen
			targetData = append(targetData, m.imageTmpBuffer[:readLen]...)
		} else {
			targetData = append(targetData, m.imageTmpBuffer[:(length-cursor)]...)
			if uint(len(targetData)) != length {
				panic("length error")
			}
			leftData = m.imageTmpBuffer[(length - cursor):readLen]
			// adbutil.Logger.Debug("leftData len: %d", len(leftData))
			// adbutil.Logger.Debug("targetData len: %d", len(targetData))
			break
		}
	}
	return
}

func (m *Minicap) saveImage(imageData []byte) {
	adbutil.Logger.Debug("save image...")
	fout, err := os.Create(fmt.Sprintf("image/image_%d.jpg", m.ImageIndex))
	defer fout.Close()
	if err != nil {
		panic(err)
	}
	fout.Write(imageData)
	m.ImageIndex += 1
}

func (m *Minicap) ReceiveImage() {
	var imageCursor uint
	var headerCursor uint
	var imageDataBuffer = make([]byte, 0)
	var imageHeaderBuffer = make([]byte, 0)
	var imageHeader []byte
	var imageData []byte
	var leftImageData []byte
	imageCursor = 0
	headerCursor = 0
	for {
		// adbutil.Logger.Debug("+++++++++++++++++++++++++++")
		// adbutil.Logger.Debug("imageCursor %d", imageCursor)
		// adbutil.Logger.Debug("headerCursor %d", headerCursor)
		imageHeader, imageData = m.readSocketDataUntilLength(ImageFrameHeaderLength - headerCursor)
		imageHeaderBuffer = append(imageHeaderBuffer, imageHeader...)
		if uint(len(imageHeaderBuffer)) != ImageFrameHeaderLength {
			adbutil.Logger.Debug("length is %d", len(imageHeaderBuffer))
			panic("image frame header length is wrong")
		}
		// rotation := m.getImageFrameRotation(imageHeaderBuffer[:ImageFrameRotationLength])
		// adbutil.Logger.Debug("rotation:%d", rotation)
		imageDataLength := m.getImageDataLength(imageHeaderBuffer[ImageFrameRotationLength:ImageFrameHeaderLength])
		// adbutil.Logger.Debug("imageDataLength: %d", imageDataLength)
		if uint(len(imageHeaderBuffer)) > ImageFrameHeaderLength {
			panic("ImageFrameHeaderLength wrong")
		}
		imageHeaderBuffer = imageHeaderBuffer[:0]
		headerCursor = 0
		if len(imageData) > 0 {
			imageDataBuffer = append(imageDataBuffer, imageData...)
			imageCursor += uint(len(imageData))
		}
		// adbutil.Logger.Debug("imageCursor:%d", imageCursor)
		// adbutil.Logger.Debug("read left socket data length: %d", imageDataLength-imageCursor)
		imageData, leftImageData = m.readSocketDataUntilLength(imageDataLength - imageCursor)
		// adbutil.Logger.Debug("imageData: %d", len(imageData))
		// adbutil.Logger.Debug("leftImageData: %d", len(leftImageData))
		imageDataBuffer = append(imageDataBuffer, imageData...)
		realImageLength := uint(len(imageDataBuffer))
		// adbutil.Logger.Debug("realImageLength: %d", realImageLength)
		if realImageLength == imageDataLength {
			// adbutil.Logger.Debug("send jpg data into chan")
			// m.saveImage(imageDataBuffer)
			m.ImageChan <- imageDataBuffer
			imageDataBuffer = imageDataBuffer[:0]
			imageCursor = 0
			leftImageDataLength := uint(len(leftImageData))
			if len(imageHeaderBuffer) != 0 {
				panic("imageHeaderBuffer is not empty")
			}
			if leftImageDataLength > 0 {
				// adbutil.Logger.Debug("leftImageDataLength: %d", leftImageDataLength)
				if leftImageDataLength <= ImageFrameHeaderLength {
					imageHeaderBuffer = append(imageHeaderBuffer, leftImageData...)
					headerCursor += leftImageDataLength
				} else {
					headerCursor += ImageFrameHeaderLength
					imageHeaderBuffer = append(imageHeaderBuffer, leftImageData[:ImageFrameHeaderLength]...)
					imageDataBuffer = append(imageDataBuffer, leftImageData[ImageFrameHeaderLength:]...)
					imageCursor += leftImageDataLength - ImageFrameHeaderLength
				}
			} else if leftImageDataLength == 0 {
				headerCursor = 0
				imageCursor = 0
			} else {
				panic("leftImageDataLength error")
			}
		} else {
			panic("imagedata length is wrong")
		}
	}
}

func (m *Minicap) Run() {
	defer m.disconnect()
	if !m.connect() {
		return
	}
	if !m.InitMinicapHeader() {
		return
	}
	m.ReceiveImage()
}
