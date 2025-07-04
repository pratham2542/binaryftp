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
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	filename := filepath.Base(filePath)
	payload := &binarygo.UploadPayload{
		FilenameLen: uint16(len(filename)),
		Filename:    []byte(filename),
		FileSize:    uint64(len(data)),
		FileData:    data,
	}

	_, err = c.send(binarygo.CMD_UPLOAD, payload)
	return err
}

func (c *Client) Download(filename string, outPath string) error {
	payload := &binarygo.DownloadPayload{
		FilenameLen: uint16(len(filename)),
		Filename:    []byte(filename),
	}

	respBytes, err := c.send(binarygo.CMD_DOWNLOAD, payload)
	if err != nil {
		return err
	}

	upload, err := binarygo.ReadUploadPayload(respBytes)
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, upload.FileData, 0644)
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

	_, err = conn.Write(headerBytes)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(payloadBytes)
	if err != nil {
		return nil, err
	}

	// Read response
	respHeader := make([]byte, binary.Size(binarygo.Header{}))
	if _, err := io.ReadFull(conn, respHeader); err != nil {
		return nil, err
	}
	headerResp, err := binarygo.ReadHeader(respHeader)
	if err != nil {
		return nil, err
	}

	payloadResp := make([]byte, headerResp.PayloadLen)
	if _, err := io.ReadFull(conn, payloadResp); err != nil {
		return nil, err
	}

	if headerResp.Status == binarygo.CMD_SUCCESS {
		return payloadResp, nil
	} else {
		msg, _ := binarygo.ReadResponseMessage(payloadResp)
		return nil, fmt.Errorf("server error: %s", string(msg.Message))
	}
}
