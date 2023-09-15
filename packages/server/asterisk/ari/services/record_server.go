package services

import (
	"agi/model"
	"github.com/pion/rtp"
	"log"
	"net"
	"os"
)

const listenAddr = "127.0.0.1:0"

type RtpServer interface {
	Address() string
	Close()
	Listen() error
}

type RtpServerImpl struct {
	address   string
	data      *model.CallMetadata
	payloadId int
	socket    net.PacketConn
	file      *os.File
}

func newRtpServer(cd *model.CallMetadata) RtpServer {
	l, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("/tmp/" + cd.Uuid + "-" + string(cd.Direction) + ".raw")

	return RtpServerImpl{
		address: l.LocalAddr().String(),
		data:    cd,
		socket:  l,
		file:    f,
	}
}

func (rtpServer RtpServerImpl) Address() string {
	return rtpServer.address
}

func (rtpServer RtpServerImpl) Close() {
	rtpServer.socket.Close()
	rtpServer.file.Close()
}

func (rtpServer RtpServerImpl) Listen() error {
	for {
		buf := make([]byte, 2000)
		packetSize, _, err := rtpServer.socket.ReadFrom(buf)
		if err != nil {
			log.Println("Error reading from socket:", err)
			return err
		}

		rtpPacket := &rtp.Packet{}
		err = rtpPacket.Unmarshal(buf[:packetSize])
		if err != nil {
			log.Println("Error unmarshalling rtp packet:", err)
			continue
		}
		rtpServer.payloadId = int(rtpPacket.PayloadType)
		_, err = rtpServer.file.Write(rtpPacket.Payload)
		if err != nil {
			log.Println("Error writing to file:", err)
		}

	}
}
