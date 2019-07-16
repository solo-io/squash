package socket

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/vishvananda/netlink/nl"
)

type inodeAndPort struct {
	inode int64
	port  int
}

func parseProcNetTcp() ([]inodeAndPort, error) {
	var sockets []inodeAndPort

	f, err := os.Open("/proc/net/tcp")
	defer f.Close()

	if err != nil {
		return nil, err
	}

	// example output:
	//   sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
	//   0: 0100007F:8A17 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 959999 1 ffff88004e4726c0 100 0 0 10 0

	tcpnetListenRegex := regexp.MustCompile(`\d+:\s+[0-9A-Fa-f]{8}:([0-9A-Fa-f]{4})\s+00000000:0000\s+0[aA]\s+00000000:00000000\s+00:00000000\s+00000000\s+\d+\s+0\s+(\d+)`)

	reader := bufio.NewReader(f)
	//read first line
	reader.ReadLine()
	for {
		l, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		if len(l) == 0 {
			continue
		}
		// now parse l:
		match := tcpnetListenRegex.FindStringSubmatch(string(l))
		if match != nil {
			portStr := match[1]
			inodeStr := match[2]

			port, err := strconv.ParseUint(portStr, 16, 16)
			if err != nil {
				continue
			}
			inode, err := strconv.ParseUint(inodeStr, 10, 64)
			if err != nil {
				continue
			}
			sockets = append(sockets, inodeAndPort{inode: int64(inode), port: int(port)})
		}
	}

	return sockets, nil
}

func GetListeningPortsFor(pid int) ([]int, error) {
	var inoddedPorts []inodeAndPort

	ctx := context.TODO()
	logger := contextutils.LoggerFrom(ctx)
	if listeningsokcets, err := SocketListen(); err != nil {
		logger.Warnw("GetListeningSocketsFor: Can't get listening sockets with netlink", "pid", pid, "err", err)

		inoddedPorts, err = parseProcNetTcp()
		if err != nil {
			logger.Errorw("GetListeningSocketsFor: Can't get listening sockets", "pid", pid, "err", err)
			return nil, err
		}
	} else {

		for _, sock := range listeningsokcets {
			curport := int(sock.ID.SourcePort)

			inoddedPorts = append(inoddedPorts, inodeAndPort{port: curport, inode: int64(sock.INode)})
		}

	}
	logger.Debugw("GetSocketInodesFor: got listening sockets", "pid", pid, "inoddedPorts", inoddedPorts)

	sockets, err := GetSocketInodesFor(pid)
	if err != nil {
		logger.Errorw("GetSocketInodesFor: Can't can socks for pid", "pid", pid, "err", err)
		return nil, err
	}
	logger.Debugw("GetSocketInodesFor: got sockets for pid", "pid", pid, "sockets", sockets)

	var pidsocks []inodeAndPort
	for _, socket := range inoddedPorts {
		for _, pidsock := range sockets {
			logger.Debugw("GetSocketInodesFor: testing socket match", "socket", socket, "inodesock", pidsock)

			if uint64(socket.inode) == pidsock {
				pidsocks = append(pidsocks, socket)
			}
		}
	}

	var ports []int
	for _, sock := range pidsocks {
		curport := int(sock.port)
		ports = append(ports, curport)
	}

	return ports, nil
}

func GetSocketInodesFor(pid int) ([]uint64, error) {
	directory := fmt.Sprintf("/proc/%d/fd/", pid)
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var inodes []uint64
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		abspath := filepath.Join(directory, f.Name())

		var stat syscall.Stat_t
		if err := syscall.Stat(abspath, &stat); err != nil {
			continue
		}
		issock := (stat.Mode & syscall.S_IFSOCK) == syscall.S_IFSOCK
		if !issock {
			continue
		}

		inodes = append(inodes, stat.Ino)
	}
	return inodes, nil
}

const (
	sizeofSocketID      = 0x30
	sizeofSocketRequest = sizeofSocketID + 0x8
	sizeofSocket        = sizeofSocketID + 0x18
)

type SocketID struct {
	SourcePort      uint16
	DestinationPort uint16
	Source          net.IP
	Destination     net.IP
	Interface       uint32
	Cookie          [2]uint32
}

// Socket represents a netlink socket.
type Socket struct {
	Family  uint8
	State   uint8
	Timer   uint8
	Retrans uint8
	ID      SocketID
	Expires uint32
	RQueue  uint32
	WQueue  uint32
	UID     uint32
	INode   uint32
}

type readBuffer struct {
	Bytes []byte
	pos   int
}

func (b *readBuffer) Read() byte {
	c := b.Bytes[b.pos]
	b.pos++
	return c
}

func (b *readBuffer) Next(n int) []byte {
	s := b.Bytes[b.pos : b.pos+n]
	b.pos += n
	return s
}

func (s *Socket) deserialize(b []byte) error {
	if len(b) < sizeofSocket {
		return fmt.Errorf("socket data short read (%d); want %d", len(b), sizeofSocket)
	}
	contextutils.LoggerFrom(context.TODO()).Debugw("Deserializing netlink socker info", "bytes", b)
	rb := readBuffer{Bytes: b}
	s.Family = rb.Read()
	s.State = rb.Read()
	s.Timer = rb.Read()
	s.Retrans = rb.Read()
	s.ID.SourcePort = networkOrder.Uint16(rb.Next(2))
	s.ID.DestinationPort = networkOrder.Uint16(rb.Next(2))

	b1 := rb.Read()
	b2 := rb.Read()
	b3 := rb.Read()
	b4 := rb.Read()
	s.ID.Source = net.IPv4(b1, b2, b3, b4)
	rb.Next(12)

	b1 = rb.Read()
	b2 = rb.Read()
	b3 = rb.Read()
	b4 = rb.Read()
	s.ID.Destination = net.IPv4(b1, b2, b3, b4)
	rb.Next(12)
	s.ID.Interface = native.Uint32(rb.Next(4))
	s.ID.Cookie[0] = native.Uint32(rb.Next(4))
	s.ID.Cookie[1] = native.Uint32(rb.Next(4))
	s.Expires = native.Uint32(rb.Next(4))
	s.RQueue = native.Uint32(rb.Next(4))
	s.WQueue = native.Uint32(rb.Next(4))
	s.UID = native.Uint32(rb.Next(4))
	s.INode = native.Uint32(rb.Next(4))
	return nil
}

type socketRequest struct {
	Family   uint8
	Protocol uint8
	Ext      uint8
	pad      uint8
	States   uint32
	ID       SocketID
}

type writeBuffer struct {
	Bytes []byte
	pos   int
}

func (b *writeBuffer) Write(c byte) {
	b.Bytes[b.pos] = c
	b.pos++
}

func (b *writeBuffer) Next(n int) []byte {
	s := b.Bytes[b.pos : b.pos+n]
	b.pos += n
	return s
}

var (
	native       = nl.NativeEndian()
	networkOrder = binary.BigEndian
)

func (r *socketRequest) Serialize() []byte {
	b := writeBuffer{Bytes: make([]byte, sizeofSocketRequest)}
	b.Write(r.Family)
	b.Write(r.Protocol)
	b.Write(r.Ext)
	b.Write(r.pad)
	native.PutUint32(b.Next(4), r.States)
	networkOrder.PutUint16(b.Next(2), r.ID.SourcePort)
	networkOrder.PutUint16(b.Next(2), r.ID.DestinationPort)
	copy(b.Next(4), r.ID.Source.To4())
	b.Next(12)
	copy(b.Next(4), r.ID.Destination.To4())
	b.Next(12)
	native.PutUint32(b.Next(4), r.ID.Interface)
	native.PutUint32(b.Next(4), r.ID.Cookie[0])
	native.PutUint32(b.Next(4), r.ID.Cookie[1])
	return b.Bytes
}

func (r *socketRequest) Len() int { return sizeofSocketRequest }

const (
	LISTEN              = 1024
	SOCK_DIAG_BY_FAMILY = 20
)

// SocketGet returns the Socket identified by its local and remote addresses.
func SocketListen() ([]*Socket, error) {

	s, err := nl.Subscribe(syscall.NETLINK_INET_DIAG)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	req := nl.NewNetlinkRequest(SOCK_DIAG_BY_FAMILY, syscall.NLM_F_REQUEST|syscall.NLM_F_DUMP)
	req.AddData(&socketRequest{
		Family:   syscall.AF_INET,
		Protocol: syscall.IPPROTO_TCP,
		States:   LISTEN,
		ID:       SocketID{},
	})
	s.Send(req)
	msgs, err := s.Receive()
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, errors.New("no message nor error from netlink")
	}
	var sockets []*Socket
	logger := contextutils.LoggerFrom(context.TODO())
	for _, msg := range msgs {
		logger.Debugw("got netlink msg", "msg", msg)
		if msg.Header.Type == SOCK_DIAG_BY_FAMILY {
			logger.Debug("got diag by family netlink msg")

			sock := &Socket{}
			if err := sock.deserialize(msg.Data); err != nil {
				logger.Warnw("Error parsing message", "err", err)
				continue
			}
			sockets = append(sockets, sock)
		} else if msg.Header.Type == syscall.NLMSG_ERROR {
			errval := native.Uint32(msg.Data[:4])
			logger.Warnw("Netlink error", "errval", errval)
			return nil, fmt.Errorf("netlink error: %d", -errval)
		}

	}
	return sockets, nil
}
