package binarygo

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// === Constants ===
const (
	// Protocol version
	PROTOCOL_VERSION = 1

	// Commands
	CMD_UPLOAD   = 1
	CMD_DOWNLOAD = 2
	CMD_LIST     = 3

	// Response types
	CMD_SUCCESS = 100
	CMD_ERROR   = 101

	MAX_FILE_SIZE = 100 * 1024 * 1024 // 100 MB

)

type PayloadEncoder interface {
	ToBytes() ([]byte, error)
}

// === Header ===
// Fixed-size header for all messages
type Header struct {
	Version    uint8  // protocol version
	Command    uint8  // command type
	Status     uint8  // status or reserved
	PayloadLen uint32 // length of the payload in bytes
}

// Serialize header to bytes
func (h *Header) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, h)
	return buf.Bytes(), err
}

// Deserialize header from bytes
func ReadHeader(data []byte) (*Header, error) {
	var h Header
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &h)
	return &h, err
}

// === Payloads ===

// --- Upload Payload ---
type UploadPayload struct {
	FilenameLen uint16
	Filename    []byte
	FileSize    uint64
	FileData    []byte
}

func (p *UploadPayload) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, p.FilenameLen); err != nil {
		return nil, err
	}
	if _, err := buf.Write(p.Filename); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.FileSize); err != nil {
		return nil, err
	}
	if _, err := buf.Write(p.FileData); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ReadUploadPayload(data []byte) (*UploadPayload, error) {
	buf := bytes.NewReader(data)
	var u UploadPayload

	if err := binary.Read(buf, binary.BigEndian, &u.FilenameLen); err != nil {
		return nil, err
	}
	u.Filename = make([]byte, u.FilenameLen)
	if _, err := buf.Read(u.Filename); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &u.FileSize); err != nil {
		return nil, err
	}
	if u.FileSize > MAX_FILE_SIZE {
		return nil, fmt.Errorf("file size %d exceeds max limit %d", u.FileSize, MAX_FILE_SIZE)
	}
	u.FileData = make([]byte, u.FileSize)
	if _, err := buf.Read(u.FileData); err != nil {
		return nil, err
	}

	return &u, nil
}

// --- Download Payload ---
type DownloadPayload struct {
	FilenameLen uint16
	Filename    []byte
}

func (p *DownloadPayload) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, p.FilenameLen)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(p.Filename)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ReadDownloadPayload(data []byte) (*DownloadPayload, error) {
	buf := bytes.NewReader(data)
	var p DownloadPayload

	if err := binary.Read(buf, binary.BigEndian, &p.FilenameLen); err != nil {
		return nil, err
	}
	p.Filename = make([]byte, p.FilenameLen)
	if _, err := buf.Read(p.Filename); err != nil {
		return nil, err
	}
	return &p, nil
}

// --- List Response Payload ---
type ListResponsePayload struct {
	FileCount uint16
	FileNames [][]byte
}

func (p *ListResponsePayload) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, p.FileCount); err != nil {
		return nil, err
	}
	for _, name := range p.FileNames {
		nameLen := uint16(len(name))
		if err := binary.Write(buf, binary.BigEndian, nameLen); err != nil {
			return nil, err
		}
		if _, err := buf.Write(name); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func ReadListResponsePayload(data []byte) (*ListResponsePayload, error) {
	buf := bytes.NewReader(data)
	var p ListResponsePayload

	if err := binary.Read(buf, binary.BigEndian, &p.FileCount); err != nil {
		return nil, err
	}

	p.FileNames = make([][]byte, 0, p.FileCount)
	for i := 0; i < int(p.FileCount); i++ {
		var nameLen uint16
		if err := binary.Read(buf, binary.BigEndian, &nameLen); err != nil {
			return nil, err
		}
		name := make([]byte, nameLen)
		if _, err := buf.Read(name); err != nil {
			return nil, err
		}
		p.FileNames = append(p.FileNames, name)
	}
	return &p, nil
}

// --- Response Message Payload ---
type ResponseMessage struct {
	MessageLen uint16
	Message    []byte
}

func (r *ResponseMessage) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, r.MessageLen)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(r.Message)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ReadResponseMessage(data []byte) (*ResponseMessage, error) {
	buf := bytes.NewReader(data)
	var r ResponseMessage

	if err := binary.Read(buf, binary.BigEndian, &r.MessageLen); err != nil {
		return nil, err
	}
	r.Message = make([]byte, r.MessageLen)
	if _, err := buf.Read(r.Message); err != nil {
		return nil, err
	}
	return &r, nil
}
