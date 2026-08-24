package jt808

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

func TestParseLocationReport(t *testing.T) {
	body := make([]byte, 28)
	binary.BigEndian.PutUint32(body[0:4], 0)
	binary.BigEndian.PutUint32(body[4:8], 3)
	binary.BigEndian.PutUint32(body[8:12], 31_123_456)
	binary.BigEndian.PutUint32(body[12:16], 121_123_456)
	binary.BigEndian.PutUint16(body[16:18], 12)
	binary.BigEndian.PutUint16(body[18:20], 600)
	binary.BigEndian.PutUint16(body[20:22], 90)
	copy(body[22:28], []byte{0x24, 0x08, 0x24, 0x12, 0x34, 0x56})
	body = append(body, 0x30, 0x01, 0x1f, 0x31, 0x01, 0x08)

	frame := makeFrame(0x0200, "013800138000", 1, body)
	result, err := New().Parse(frame)
	if err != nil {
		t.Fatal(err)
	}
	if result.Protocol != "JT-808" || result.MessageID != "0200" || result.MessageName != "位置信息汇报" {
		t.Fatalf("unexpected result header: %+v", result)
	}
	data := result.Data.(map[string]interface{})
	header := data["header"].(map[string]interface{})
	if header["terminalPhone"] != "13800138000" {
		t.Fatalf("unexpected normalized terminal phone: %v", header["terminalPhone"])
	}
	if math.Abs(data["latitude"].(float64)-31.123456) > 0.000001 || math.Abs(data["longitude"].(float64)-121.123456) > 0.000001 {
		t.Fatalf("unexpected coordinates: %#v", data)
	}
	if data["locationTime"] != "2024-08-24 12:34:56（北京时间）" {
		t.Fatalf("unexpected time: %v", data["locationTime"])
	}
	if len(result.Fields) < 18 {
		t.Fatalf("expected field-level output, got %d fields", len(result.Fields))
	}
	for _, field := range result.Fields {
		if (field.Name == "纬度" || field.Name == "经度") && strings.Contains(field.Value, "°") {
			t.Fatalf("coordinate value must remain numeric: %s=%s", field.Name, field.Value)
		}
	}
}

func TestNormalizePhoneRemovesOnlyProtocolPadding(t *testing.T) {
	if value := normalizePhone("016180056163"); value != "16180056163" {
		t.Fatalf("unexpected mainland phone: %s", value)
	}
	if value := normalizePhone("001380013800"); value != "01380013800" {
		t.Fatalf("only one padding zero should be removed: %s", value)
	}
	if value := normalizePhone("886912345678"); value != "886912345678" {
		t.Fatalf("12-digit non-mainland number must be preserved: %s", value)
	}
}

func TestEscapedTextMessage(t *testing.T) {
	frame := makeFrame(0x8300, "013800138000", 2, []byte{0x00, 0x7e, 0x7d})
	if !strings.Contains(strings.ToLower(stringHex(frame)), "7d02") || !strings.Contains(strings.ToLower(stringHex(frame)), "7d01") {
		t.Fatalf("test frame was not escaped: %x", frame)
	}
	result, err := New().Parse(frame)
	if err != nil {
		t.Fatal(err)
	}
	data := result.Data.(map[string]interface{})
	if data["text"] != "~}" {
		t.Fatalf("unexpected decoded text: %#v", data)
	}
}

func TestRejectsInvalidChecksum(t *testing.T) {
	frame := makeFrame(0x0002, "013800138000", 3, nil)
	frame[len(frame)-2] ^= 0x01
	_, err := New().Parse(frame)
	if err == nil || !strings.Contains(err.Error(), "校验失败") {
		t.Fatalf("expected checksum error, got %v", err)
	}
}

func makeFrame(messageID uint16, phone string, serial uint16, body []byte) []byte {
	packet := make([]byte, 12+len(body)+1)
	binary.BigEndian.PutUint16(packet[0:2], messageID)
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(body)))
	for index := 0; index < 6; index++ {
		packet[4+index] = (phone[index*2]-'0')<<4 | (phone[index*2+1] - '0')
	}
	binary.BigEndian.PutUint16(packet[10:12], serial)
	copy(packet[12:], body)
	for _, value := range packet[:len(packet)-1] {
		packet[len(packet)-1] ^= value
	}
	frame := []byte{0x7e}
	for _, value := range packet {
		switch value {
		case 0x7e:
			frame = append(frame, 0x7d, 0x02)
		case 0x7d:
			frame = append(frame, 0x7d, 0x01)
		default:
			frame = append(frame, value)
		}
	}
	return append(frame, 0x7e)
}

func stringHex(data []byte) string {
	const chars = "0123456789abcdef"
	result := make([]byte, len(data)*2)
	for index, value := range data {
		result[index*2] = chars[value>>4]
		result[index*2+1] = chars[value&0x0f]
	}
	return string(result)
}
