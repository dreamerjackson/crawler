package generator

import (
	"bytes"
	"encoding/binary"
	"net"
)

func IDbyIP(ip string) uint32 {
	var id uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &id)
	return id
}
