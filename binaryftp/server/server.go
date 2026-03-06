package binaryftp

import (
	binarygo "binary-go/binary-cust"
	"encoding/binary"
	"io"
	"log"
	"net"
)

type Storage interface {
	Save(name string, r io.Reader, size uint64) error
	Get(name string) (io.ReadCloser, uint64, error)
	List() ([]string, error)
}

type Server struct {
	addr    string
	storage Storage
}

func New(addr string, storage Storage) *Server {
	return &Server{
		addr:    addr,
		storage: storage,
	}
}

func (s *Server) Start() error {

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Println("server listening on", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	headerBuf := make([]byte, binary.Size(binarygo.Header{}))

	if _, err := io.ReadFull(conn, headerBuf); err != nil {
		log.Println("header read error:", err)
		return
	}

	header, err := binarygo.ReadHeader(headerBuf)
	if err != nil {
		log.Println("header parse error:", err)
		return
	}

	switch header.Command {

	case binarygo.CMD_UPLOAD:
		if err := s.handleUpload(conn, header); err != nil {
			sendError(conn, err.Error())
		}

	case binarygo.CMD_DOWNLOAD:
		if err := s.handleDownload(conn, header); err != nil {
			sendError(conn, err.Error())
		}

	case binarygo.CMD_LIST:
		if err := s.handleList(conn, header); err != nil {
			sendError(conn, err.Error())
		}

	default:
		sendError(conn, "unknown command")
	}
}

func (s *Server) handleUpload(conn net.Conn, header *binarygo.Header) error {

	payload := make([]byte, header.PayloadLen)

	if _, err := io.ReadFull(conn, payload); err != nil {
		return err
	}

	meta, err := binarygo.ReadUploadPayload(payload)
	if err != nil {
		return err
	}

	reader := io.LimitReader(conn, int64(meta.FileSize))

	err = s.storage.Save(string(meta.Filename), reader, meta.FileSize)
	if err != nil {
		return err
	}

	sendSuccess(conn, nil)
	return nil
}

func (s *Server) handleDownload(conn net.Conn, header *binarygo.Header) error {

	payload := make([]byte, header.PayloadLen)

	if _, err := io.ReadFull(conn, payload); err != nil {
		return err
	}

	dp, err := binarygo.ReadDownloadPayload(payload)
	if err != nil {
		return err
	}

	reader, size, err := s.storage.Get(string(dp.Filename))
	if err != nil {
		return err
	}
	defer reader.Close()

	filename := string(dp.Filename)

	meta := &binarygo.UploadPayload{
		FilenameLen: uint16(len(filename)),
		Filename:    []byte(filename),
		FileSize:    size,
	}

	metaBytes, err := meta.ToBytes()
	if err != nil {
		return err
	}

	respHeader := &binarygo.Header{
		Version:    binarygo.PROTOCOL_VERSION,
		Command:    0,
		Status:     binarygo.CMD_SUCCESS,
		PayloadLen: uint32(len(metaBytes)),
	}

	hdr, err := respHeader.ToBytes()
	if err != nil {
		return err
	}

	if _, err := conn.Write(hdr); err != nil {
		return err
	}

	if _, err := conn.Write(metaBytes); err != nil {
		return err
	}

	if _, err := io.Copy(conn, reader); err != nil {
		return err
	}

	return nil
}

func (s *Server) handleList(conn net.Conn, header *binarygo.Header) error {

	files, err := s.storage.List()
	if err != nil {
		return err
	}

	payload := &binarygo.ListResponsePayload{
		FileCount: uint16(len(files)),
		FileNames: make([][]byte, 0, len(files)),
	}

	for _, f := range files {
		payload.FileNames = append(payload.FileNames, []byte(f))
	}

	bytes, err := payload.ToBytes()
	if err != nil {
		return err
	}

	sendSuccess(conn, bytes)
	return nil
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
