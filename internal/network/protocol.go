package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const maxMsgSize = 500 * 1024 * 1024 // 500 МБ

type Message struct {
	MIME    string
	Payload []byte
}

// WriteMessage пишет сообщение в бинарном формате:
// [totalLen:4][mimeLen:4][mime][payloadLen:4][payload]
func WriteMessage(conn net.Conn, msg Message) error {
	mimeBytes := []byte(msg.MIME)

	// Считаем общую длину тела
	totalLen := uint32(4 + len(mimeBytes) + 4 + len(msg.Payload))

	// Пишем в буфер по порядку
	buf := make([]byte, 4+totalLen)
	offset := 0

	// 1. Общая длина
	binary.BigEndian.PutUint32(buf[offset:offset+4], totalLen)
	offset += 4

	// 2. MIME
	binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(len(mimeBytes)))
	offset += 4
	copy(buf[offset:offset+len(mimeBytes)], mimeBytes)
	offset += len(mimeBytes)

	// 3. Payload
	binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(len(msg.Payload)))
	offset += 4
	copy(buf[offset:], msg.Payload)

	_, err := conn.Write(buf)
	return err
}

// ReadMessage читает одно сообщение из conn
func ReadMessage(conn net.Conn) (Message, error) {
	// 1. Читаем общую длину
	var totalLen uint32
	if err := binary.Read(conn, binary.BigEndian, &totalLen); err != nil {
		return Message{}, err
	}
	if totalLen > maxMsgSize {
		return Message{}, fmt.Errorf("message too large: %d bytes", totalLen)
	}
	if totalLen < 8 { // минимум: 2 поля по 4 байта (длины)
		return Message{}, fmt.Errorf("message too small: %d bytes", totalLen)
	}

	// 2. Читаем всё тело
	body := make([]byte, totalLen)
	if _, err := io.ReadFull(conn, body); err != nil {
		return Message{}, err
	}

	offset := 0

	// 3. MIME
	mimeLen := int(binary.BigEndian.Uint32(body[offset : offset+4]))
	offset += 4
	if offset+mimeLen > len(body) {
		return Message{}, fmt.Errorf("invalid mime length")
	}
	mime := string(body[offset : offset+mimeLen])
	offset += mimeLen

	// 4. Payload
	if offset+4 > len(body) {
		return Message{}, fmt.Errorf("truncated message")
	}
	payloadLen := int(binary.BigEndian.Uint32(body[offset : offset+4]))
	offset += 4
	if offset+payloadLen != len(body) {
		return Message{}, fmt.Errorf("payload size mismatch: expected %d, got %d", payloadLen, len(body)-offset)
	}
	payload := body[offset:]

	return Message{
		MIME:    mime,
		Payload: payload,
	}, nil
}
