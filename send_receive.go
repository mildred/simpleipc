package simpleipc

import (
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"unsafe"
)

func ToFd(files []*os.File) []int {
	fds := make([]int, len(files))
	for i := range files {
		fds[i] = int(files[i].Fd())
	}
	return fds
}

func Send(cnx *net.UnixConn, bytes []byte, fds []int) error {
	viaf, err := cnx.File()
	if err != nil {
		return fmt.Errorf("convert connection to file descriptor: %v", err)
	}
	socket := int(viaf.Fd())
	defer viaf.Close()

	rights := syscall.UnixRights(fds...)
	return syscall.Sendmsg(socket, bytes, rights, nil, 0)
}

type Response struct {
	Data []byte
	cmsg []byte
	Fds  []int
}

func Receive(cnx *net.UnixConn, length int, maxfiles int) (res Response, err error) {
	// get the underlying socket
	viaf, err := cnx.File()
	if err != nil {
		return res, fmt.Errorf("convert connection to file descriptor: %v", err)
	}
	socket := int(viaf.Fd())
	defer viaf.Close()

	// recvmsg
	res.Data = make([]byte, length)
	res.cmsg = make([]byte, syscall.CmsgSpace(maxfiles*4))
	var nres int
	nres, _, _, _, err = syscall.Recvmsg(socket, res.Data, res.cmsg, 0)
	if err != nil {
		return res, err
	} else if nres == 0 {
		return res, io.EOF
	}
	return res, nil
}

func (res *Response) ParseFiles(numFiles int) error {
	var err error
	res.Fds = []int{}
	if numFiles == 0 {
		return nil
	}

	res.cmsg = res.cmsg[:syscall.CmsgSpace(numFiles*4)]

	// parse control msgs
	var msgs []syscall.SocketControlMessage
	msgs, err = syscall.ParseSocketControlMessage(res.cmsg)
	if err != nil {
		h := (*syscall.Cmsghdr)(unsafe.Pointer(&res.cmsg[0]))
		fmt.Printf("len(b): %d\n", len(res.cmsg))
		fmt.Printf("SizeofCmsghdr: %d\n", syscall.SizeofCmsghdr)
		fmt.Printf("h.Len: %d\n", h.Len)
		return fmt.Errorf("parse control message: %v", err)
	}

	for i := 0; i < len(msgs) && err == nil; i++ {
		fds, err := syscall.ParseUnixRights(&msgs[i])
		if err != nil {
			return fmt.Errorf("parse file descriptors: %v", err)
		}
		res.Fds = append(res.Fds, fds...)
	}

	return nil
}

func ToFiles(fds []int, filenames []string) []*os.File {
	res := make([]*os.File, 0, len(fds))
	for fi, fd := range fds {
		var filename string
		if fi < len(filenames) {
			filename = filenames[fi]
		}
		res = append(res, os.NewFile(uintptr(fd), filename))
	}
	return res
}
