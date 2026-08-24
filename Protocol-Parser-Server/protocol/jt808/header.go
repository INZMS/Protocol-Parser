package jt808

import (
	"encoding/hex"
	"fmt"

	"protocol-parser-server/parser/core"
)

type Header struct {
	MessageID     uint16
	Properties    uint16
	BodyLength    int
	Encryption    int
	Fragmented    bool
	Phone         string
	Serial        uint16
	TotalPackages uint16
	PackageIndex  uint16
	Length        int
}

func parseHeader(packet []byte) (*Header, error) {
	if len(packet) < 13 {
		return nil, fmt.Errorf("JT-808消息头长度不足")
	}
	properties := uint16(packet[2])<<8 | uint16(packet[3])
	header := &Header{
		MessageID:  uint16(packet[0])<<8 | uint16(packet[1]),
		Properties: properties,
		BodyLength: int(properties & 0x03ff),
		Encryption: int((properties >> 10) & 0x07),
		Fragmented: properties&(1<<13) != 0,
		Phone:      normalizePhone(decodeBCDLoose(packet[4:10])),
		Serial:     uint16(packet[10])<<8 | uint16(packet[11]),
		Length:     12,
	}
	if header.Fragmented {
		if len(packet) < 17 {
			return nil, fmt.Errorf("JT-808分包消息缺少消息包封装项")
		}
		header.TotalPackages = uint16(packet[12])<<8 | uint16(packet[13])
		header.PackageIndex = uint16(packet[14])<<8 | uint16(packet[15])
		header.Length = 16
		if header.TotalPackages == 0 || header.PackageIndex == 0 || header.PackageIndex > header.TotalPackages {
			return nil, fmt.Errorf("JT-808分包信息无效: 总包数%d，包序号%d", header.TotalPackages, header.PackageIndex)
		}
	}
	return header, nil
}

func headerFields(header *Header, packet []byte) []core.Field {
	fields := []core.Field{
		newField(1, "起始标识", 0, []byte{0x7e}, "0x7E", "报文开始标识"),
		newField(2, "消息ID", 1, packet[0:2], fmt.Sprintf("0x%04X（%s）", header.MessageID, messageName(header.MessageID)), "WORD，大端字节序"),
		newField(3, "消息体属性", 3, packet[2:4], fmt.Sprintf("0x%04X", header.Properties), "bit0-9长度，bit10-12加密，bit13分包"),
		newField(4, "消息体长度", 3, packet[2:4], fmt.Sprintf("%d Bytes", header.BodyLength), "消息体属性bit0-bit9"),
		newField(5, "数据加密方式", 3, packet[2:4], encryptionText(header.Encryption), "0不加密，1为RSA，其余保留"),
		newField(6, "分包标识", 3, packet[2:4], boolText(header.Fragmented, "分包消息", "非分包消息"), "消息体属性bit13"),
		newField(7, "终端手机号", 5, packet[4:10], header.Phone, "BCD[6]，不足12位前补0"),
		newField(8, "消息流水号", 11, packet[10:12], fmt.Sprint(header.Serial), "WORD，发送顺序循环累加"),
	}
	if header.Fragmented {
		fields = append(fields,
			newField(9, "消息总包数", 13, packet[12:14], fmt.Sprint(header.TotalPackages), "分包后的总包数"),
			newField(10, "包序号", 15, packet[14:16], fmt.Sprint(header.PackageIndex), "从1开始"),
		)
	}
	return fields
}

func newField(index int, name string, offset int, raw []byte, value, description string) core.Field {
	return core.Field{Index: index, Name: name, Offset: offset, Length: len(raw), Raw: hex.EncodeToString(raw), Value: value, Description: description}
}

func encryptionText(value int) string {
	switch value {
	case 0:
		return "不加密（0）"
	case 1:
		return "RSA加密（1）"
	default:
		return fmt.Sprintf("保留加密类型（%d）", value)
	}
}

func boolText(value bool, yes, no string) string {
	if value {
		return yes
	}
	return no
}
