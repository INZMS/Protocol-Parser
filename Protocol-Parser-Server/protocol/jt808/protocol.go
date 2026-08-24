package jt808

import (
	"encoding/hex"
	"fmt"

	"protocol-parser-server/parser/core"
)

type ProtocolJT808 struct{}

func New() core.Protocol { return &ProtocolJT808{} }

func (p *ProtocolJT808) Name() string { return "JT-808" }

func (p *ProtocolJT808) Match(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x7e && data[len(data)-1] == 0x7e
}

func (p *ProtocolJT808) Parse(data []byte) (*core.ParseResult, error) {
	packet, err := decodeFrame(data)
	if err != nil {
		return nil, err
	}
	header, err := parseHeader(packet)
	if err != nil {
		return nil, err
	}
	expectedLength := header.Length + header.BodyLength + 1
	if len(packet) != expectedLength {
		return nil, fmt.Errorf("JT-808报文长度不一致: 消息头声明%d字节消息体，实际解码后包长%d字节", header.BodyLength, len(packet))
	}

	fields := headerFields(header, packet)
	body := packet[header.Length : header.Length+header.BodyLength]
	bodyData, err := parseBody(header, body, &fields)
	if err != nil {
		return nil, err
	}
	fields = append(fields,
		newField(len(fields)+1, "校验码", 1+len(packet)-1, packet[len(packet)-1:], fmt.Sprintf("0x%02X", packet[len(packet)-1]), "消息头至消息体末字节逐字节异或"),
		newField(len(fields)+2, "结束标识", 1+len(packet), []byte{0x7e}, "0x7E", "报文结束标识"),
	)

	dataResult := bodyData
	if dataResult == nil {
		dataResult = map[string]interface{}{}
	}
	dataResult["header"] = map[string]interface{}{
		"terminalPhone":  header.Phone,
		"serialNumber":   header.Serial,
		"bodyLength":     header.BodyLength,
		"encryptionType": header.Encryption,
		"fragmented":     header.Fragmented,
		"totalPackages":  header.TotalPackages,
		"packageIndex":   header.PackageIndex,
	}

	return &core.ParseResult{
		Protocol:    p.Name(),
		MessageID:   fmt.Sprintf("%04x", header.MessageID),
		MessageName: messageName(header.MessageID),
		Length:      len(data),
		Raw:         hex.EncodeToString(data),
		Fields:      fields,
		Data:        dataResult,
	}, nil
}

func decodeFrame(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0x7e || data[len(data)-1] != 0x7e {
		return nil, fmt.Errorf("JT-808报文必须以0x7E开始并以0x7E结束")
	}
	decoded := make([]byte, 0, len(data)-2)
	for index := 1; index < len(data)-1; index++ {
		if data[index] != 0x7d {
			decoded = append(decoded, data[index])
			continue
		}
		if index+1 >= len(data)-1 {
			return nil, fmt.Errorf("JT-808转义序列不完整")
		}
		index++
		switch data[index] {
		case 0x01:
			decoded = append(decoded, 0x7d)
		case 0x02:
			decoded = append(decoded, 0x7e)
		default:
			return nil, fmt.Errorf("JT-808未知转义序列: 7D%02X", data[index])
		}
	}
	if len(decoded) < 13 {
		return nil, fmt.Errorf("JT-808报文过短: 解码后至少需要13字节")
	}
	var checksum byte
	for _, value := range decoded[:len(decoded)-1] {
		checksum ^= value
	}
	if checksum != decoded[len(decoded)-1] {
		return nil, fmt.Errorf("JT-808校验失败: 计算值0x%02X，报文值0x%02X", checksum, decoded[len(decoded)-1])
	}
	return decoded, nil
}
