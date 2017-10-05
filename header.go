package simpleipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

func (h *Header) WriteWithPayload(cnx *net.UnixConn, payload []byte) error {
	if uint32(len(payload)) != h.Size {
		return fmt.Errorf("Payload size (%d) in header different than actual payload (%d)", h.Size, len(payload))
	}
	err := h.Write(cnx)
	if err != nil {
		return err
	}
	_, err = cnx.Write(payload)
	return err
}

func (h *Header) read(cnx *net.UnixConn, filenames []string, inclPayload bool) ([]byte, error) {
	var payload []byte
	res, err := Receive(cnx, HeaderLength, HeaderMaxFiles)
	if err != nil {
		return payload, err
	}
	h.Decode(res.Data)
	if inclPayload && h.Size > 0 {
		payload = make([]byte, h.Size)
		_, err = io.ReadFull(cnx, payload)
		if err != nil {
			return payload, err
		}
	}
	err = res.ParseFiles(int(h.NumFiles))
	if err != nil {
		return payload, err
	}
	h.Files = ToFiles(res.Fds, filenames)
	return payload, nil
}

func (h *Header) Read(cnx *net.UnixConn, filenames []string) error {
	_, err := h.read(cnx, filenames, false)
	return err
}

func (h *Header) ReadWithPayload(cnx *net.UnixConn, filenames []string) ([]byte, error) {
	return h.read(cnx, filenames, true)
}
