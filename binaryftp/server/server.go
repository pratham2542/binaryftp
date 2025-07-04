package binaryftp

import (
	binarygo "binary-go/binary-cust"
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	addr     string
	listener net.Listener

	onUpload   func(filename string, data []byte) error
	onDownload func(filename string) ([]byte, error)
	onList     func() ([]string, error)
}

func New(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) OnUpload(handler func(filename string, data []byte) error) {
	s.onUpload = handler
}

func (s *Server) OnDownload(handler func(filename string) ([]byte, error)) {
	s.onDownload = handler
}

func (s *Server) OnList(handler func() ([]string, error)) {
	s.onList = handler
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln
	log.Println("Server listening on", s.addr)

	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConnection(conn)
		}()
	}
}
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	headerBuf := make([]byte, binary.Size(binarygo.Header{}))
	if _, err := io.ReadFull(conn, headerBuf); err != nil {
		return
	}

	header, err := binarygo.ReadHeader(headerBuf)
	if err != nil {
		return
	}

	payload := make([]byte, header.PayloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return
	}

	switch header.Command {
	case binarygo.CMD_UPLOAD:
		s.handleUpload(conn, header, payload)
	case binarygo.CMD_DOWNLOAD:
		s.handleDownload(conn, header, payload)
	case binarygo.CMD_LIST:
		s.handleList(conn, header)
	default:
		sendError(conn, "Unknown command")
	}
}

func (s *Server) handleUpload(conn net.Conn, header *binarygo.Header, payload []byte) {
	up, err := binarygo.ReadUploadPayload(payload)
	if err != nil {
		sendError(conn, "Invalid upload payload")
		return
	}
	if s.onUpload != nil {
		err = s.onUpload(string(up.Filename), up.FileData)
		if err != nil {
			sendError(conn, err.Error())
			return
		}
	}
	sendSuccess(conn, nil)
}

func (s *Server) handleDownload(conn net.Conn, header *binarygo.Header, payload []byte) {
	dp, err := binarygo.ReadDownloadPayload(payload)
	if err != nil {
		sendError(conn, "Invalid download payload")
		return
	}
	if s.onDownload == nil {
		sendError(conn, "No download handler registered")
		return
	}
	data, err := s.onDownload(string(dp.Filename))
	if err != nil {
		sendError(conn, err.Error())
		return
	}

	up := &binarygo.UploadPayload{
		FilenameLen: dp.FilenameLen,
		Filename:    dp.Filename,
		FileSize:    uint64(len(data)),
		FileData:    data,
	}
	bytes, _ := up.ToBytes()
	sendSuccess(conn, bytes)
}

func (s *Server) handleList(conn net.Conn, header *binarygo.Header) {
	if s.onList == nil {
		sendError(conn, "No list handler registered")
		return
	}
	names, err := s.onList()
	if err != nil {
		sendError(conn, err.Error())
		return
	}
	p := &binarygo.ListResponsePayload{
		FileCount: uint16(len(names)),
		FileNames: make([][]byte, 0, len(names)),
	}
	for _, name := range names {
		p.FileNames = append(p.FileNames, []byte(name))
	}
	bytes, _ := p.ToBytes()
	sendSuccess(conn, bytes)
}

func sendSuccess(conn net.Conn, payload []byte) {
	header := &binarygo.Header{
		Version:    binarygo.PROTOCOL_VERSION,
		Command:    0,
		Status:     binarygo.CMD_SUCCESS,
		PayloadLen: uint32(len(payload)),
	}
	hdr, _ := header.ToBytes()
	conn.Write(hdr)
	conn.Write(payload)
}

func sendError(conn net.Conn, msg string) {
	resp := &binarygo.ResponseMessage{
		MessageLen: uint16(len(msg)),
		Message:    []byte(msg),
	}
	bytes, _ := resp.ToBytes()

	header := &binarygo.Header{
		Version:    binarygo.PROTOCOL_VERSION,
		Command:    0,
		Status:     binarygo.CMD_ERROR,
		PayloadLen: uint32(len(bytes)),
	}
	hdr, _ := header.ToBytes()
	conn.Write(hdr)
	conn.Write(bytes)
}
