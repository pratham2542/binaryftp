package binarygo

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestHeaderSerialization(t *testing.T) {
	header := &Header{
		Version:    PROTOCOL_VERSION,
		Command:    CMD_UPLOAD,
		Status:     0,
		PayloadLen: 12345,
	}

	data, err := header.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	deserialized, err := ReadHeader(data)
	if err != nil {
		t.Fatal(err)
	}

	if *header != *deserialized {
		t.Errorf("expected %+v, got %+v", header, deserialized)
	}
}

func TestUploadPayloadRoundTrip(t *testing.T) {
	original := &UploadPayload{
		FilenameLen: 8,
		Filename:    []byte("test.txt"),
		FileSize:    6,
	}

	data, err := original.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ReadUploadPayload(data)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(original.Filename, parsed.Filename) ||
		original.FileSize != parsed.FileSize {
		t.Errorf("upload payload mismatch")
	}
}

func TestUploadPayloadMaxFileSize(t *testing.T) {
	oversize := MAX_FILE_SIZE + 1
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(4))
	buf.Write([]byte("file"))
	binary.Write(buf, binary.BigEndian, uint64(oversize))
	buf.Write(make([]byte, 1))

	_, err := ReadUploadPayload(buf.Bytes())
	if err == nil {
		t.Fatal("expected error for oversized file, got none")
	}
}

func TestDownloadPayloadRoundTrip(t *testing.T) {
	original := &DownloadPayload{
		FilenameLen: 8,
		Filename:    []byte("file.txt"),
	}

	data, err := original.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ReadDownloadPayload(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original.Filename, parsed.Filename) {
		t.Errorf("download payload mismatch")
	}
}

func TestListResponsePayloadRoundTrip(t *testing.T) {
	names := [][]byte{
		[]byte("file1.txt"),
		[]byte("file2.log"),
	}
	original := &ListResponsePayload{
		FileCount: 2,
		FileNames: names,
	}

	data, err := original.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ReadListResponsePayload(data)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.FileCount != original.FileCount {
		t.Errorf("file count mismatch")
	}

	for i := range names {
		if !bytes.Equal(parsed.FileNames[i], names[i]) {
			t.Errorf("filename %d mismatch", i)
		}
	}
}

func TestResponseMessageRoundTrip(t *testing.T) {
	msg := &ResponseMessage{
		MessageLen: 14,
		Message:    []byte("hello, client!"),
	}

	data, err := msg.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ReadResponseMessage(data)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(msg.Message, parsed.Message) {
		t.Errorf("response message mismatch")
	}
}
