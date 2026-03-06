package binaryftp

import (
	binarygo "binary-go/binary-cust"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

type Client struct {
	Addr string
}

func New(addr string) *Client {
	return &Client{Addr: addr}
}

func (c *Client) Upload(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	filename := filepath.Base(filePath)

	meta := &binarygo.UploadPayload{
		FilenameLen: uint16(len(filename)),
		Filename:    []byte(filename),
		FileSize:    uint64(info.Size()),
	}

	metaBytes, err := meta.ToBytes()
	if err != nil {
		return err
	}

	header := &binarygo.Header{
		Version:    binarygo.PROTOCOL_VERSION,
		Command:    binarygo.CMD_UPLOAD,
		Status:     0,
		PayloadLen: uint32(len(metaBytes)),
	}

	headerBytes, err := header.ToBytes()
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Write(headerBytes); err != nil {
		return err
	}

	if _, err := conn.Write(metaBytes); err != nil {
		return err
	}

	if _, err := io.Copy(conn, file); err != nil {
		return err
	}

	respHeaderBuf := make([]byte, binary.Size(binarygo.Header{}))

	if _, err := io.ReadFull(conn, respHeaderBuf); err != nil {
		return err
	}

	respHeader, err := binarygo.ReadHeader(respHeaderBuf)
	if err != nil {
		return err
	}

	if respHeader.Status == binarygo.CMD_ERROR {

		errBuf := make([]byte, respHeader.PayloadLen)

		if _, err := io.ReadFull(conn, errBuf); err != nil {
			return err
		}

		msg, err := binarygo.ReadResponseMessage(errBuf)
		if err != nil {
			return err
		}
		return fmt.Errorf("server error: %s", string(msg.Message))
	}

	return nil
}

func (c *Client) Download(filename string, outPath string) error {

	payload := &binarygo.DownloadPayload{
		FilenameLen: uint16(len(filename)),
		Filename:    []byte(filename),
	}

	payloadBytes, err := payload.ToBytes()
	if err != nil {
		return err
	}

	header := &binarygo.Header{
		Version:    binarygo.PROTOCOL_VERSION,
		Command:    binarygo.CMD_DOWNLOAD,
		Status:     0,
		PayloadLen: uint32(len(payloadBytes)),
	}

	headerBytes, err := header.ToBytes()
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Write(headerBytes); err != nil {
		return err
	}

	if _, err := conn.Write(payloadBytes); err != nil {
		return err
	}

	respHeaderBuf := make([]byte, binary.Size(binarygo.Header{}))

	if _, err := io.ReadFull(conn, respHeaderBuf); err != nil {
		return err
	}

	respHeader, err := binarygo.ReadHeader(respHeaderBuf)
	if err != nil {
		return err
	}

	if respHeader.Status == binarygo.CMD_ERROR {

		errBuf := make([]byte, respHeader.PayloadLen)

		if _, err := io.ReadFull(conn, errBuf); err != nil {
			return err
		}

		msg, _ := binarygo.ReadResponseMessage(errBuf)

		return fmt.Errorf("server error: %s", string(msg.Message))
	}

	metaBuf := make([]byte, respHeader.PayloadLen)

	if _, err := io.ReadFull(conn, metaBuf); err != nil {
		return err
	}

	meta, err := binarygo.ReadUploadPayload(metaBuf)
	if err != nil {
		return err
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.CopyN(out, conn, int64(meta.FileSize))
	return err
}

func (c *Client) ListFiles() ([]string, error) {

	respBytes, err := c.send(binarygo.CMD_LIST, nil)
	if err != nil {
		return nil, err
	}

	list, err := binarygo.ReadListResponsePayload(respBytes)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, list.FileCount)

	for _, name := range list.FileNames {
		names = append(names, string(name))
	}

	return names, nil
}

func (c *Client) send(cmd uint8, payload binarygo.PayloadEncoder) ([]byte, error) {

	payloadBytes := []byte{}
	var err error

	if payload != nil {
		payloadBytes, err = payload.ToBytes()
		if err != nil {
			return nil, err
		}
	}

	header := &binarygo.Header{
		Version:    binarygo.PROTOCOL_VERSION,
		Command:    cmd,
		Status:     0,
		PayloadLen: uint32(len(payloadBytes)),
	}

	headerBytes, err := header.ToBytes()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if _, err := conn.Write(headerBytes); err != nil {
		return nil, err
	}

	if _, err := conn.Write(payloadBytes); err != nil {
		return nil, err
	}

	respHeaderBuf := make([]byte, binary.Size(binarygo.Header{}))

	if _, err := io.ReadFull(conn, respHeaderBuf); err != nil {
		return nil, err
	}

	respHeader, err := binarygo.ReadHeader(respHeaderBuf)
	if err != nil {
		return nil, err
	}

	payloadResp := make([]byte, respHeader.PayloadLen)

	if _, err := io.ReadFull(conn, payloadResp); err != nil {
		return nil, err
	}

	if respHeader.Status == binarygo.CMD_SUCCESS {
		return payloadResp, nil
	}

	msg, err := binarygo.ReadResponseMessage(payloadResp)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("server error: %s", string(msg.Message))
}
