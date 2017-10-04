package simpleipc

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
)

const HeaderLength = 12
const HeaderMaxFiles = 256

type Header struct {
	Seq       uint32
	Reserved0 uint8
	Reserved1 uint16
	NumFiles  uint8
	Size      uint32
	Files     []*os.File
}

func NewHeader(seq uint32, size uint32, files []*os.File) *Header {
	return &Header{
		Seq:       seq,
		Reserved0: 0,
		Reserved1: 0,
		NumFiles:  uint8(len(files)),
		Size:      size,
		Files:     files,
	}
}

func (h *Header) Encode() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, h.Seq)
	binary.Write(&buf, binary.BigEndian, h.Reserved0)
	binary.Write(&buf, binary.BigEndian, h.Reserved1)
	binary.Write(&buf, binary.BigEndian, h.NumFiles)
	binary.Write(&buf, binary.BigEndian, h.Size)
	return buf.Bytes()
}

func (h *Header) Decode(data []byte) {
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &h.Seq)
	binary.Read(buf, binary.BigEndian, &h.Reserved0)
	binary.Read(buf, binary.BigEndian, &h.Reserved1)
	binary.Read(buf, binary.BigEndian, &h.NumFiles)
	binary.Read(buf, binary.BigEndian, &h.Size)
}

func (h *Header) Write(cnx *net.UnixConn) error {
	data := h.Encode()
	fds := ToFd(h.Files)
	return Send(cnx, data, fds)
}

func (h *Header) Read(cnx *net.UnixConn, filenames []string) error {
	res, err := Receive(cnx, HeaderLength, HeaderMaxFiles)
	if err != nil {
		return err
	}
	h.Decode(res.Data)
	err = res.ParseFiles(int(h.NumFiles))
	if err != nil {
		return err
	}
	h.Files = ToFiles(res.Fds, filenames)
	return nil
}
