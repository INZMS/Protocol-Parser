package vdf

import (
	"encoding/hex"
	"fmt"
	"strings"

	"protocol-parser-server/parser/core"
)

type ProtocolVDF struct{}

func New() core.Protocol { return &ProtocolVDF{} }

func (p *ProtocolVDF) Name() string { return "VDF" }

func (p *ProtocolVDF) Match(data []byte) bool {
	return len(data) >= 10 && strings.HasPrefix(string(data), "*HQ20") && data[len(data)-1] == '#'
}

func (p *ProtocolVDF) Parse(data []byte) (*core.ParseResult, error) {
	if !p.Match(data) {
		return nil, fmt.Errorf("VDF报文必须以*HQ20开始并以#结束")
	}
	for index, value := range data {
		if value < 0x20 || value > 0x7e {
			return nil, fmt.Errorf("VDF普通报文应为ASCII字符，偏移%d处为0x%02X", index, value)
		}
	}
	comma := strings.IndexByte(string(data[6:len(data)-1]), ',')
	if comma < 0 {
		return nil, fmt.Errorf("VDF报文缺少终端ID后的逗号分隔符")
	}
	comma += 6
	if comma+3 > len(data)-1 {
		return nil, fmt.Errorf("VDF报文缺少功能类型或功能项关键字")
	}

	reply := data[5]
	terminalID := string(data[6:comma])
	if reply != '0' && reply != '1' {
		return nil, fmt.Errorf("VDF回复属性必须为0或1，实际为%c", reply)
	}
	if terminalID == "" {
		return nil, fmt.Errorf("VDF终端ID不能为空")
	}
	functionType := data[comma+1]
	keyword := data[comma+2]
	messageID := string([]byte{functionType, keyword})
	payloadStart := comma + 3
	payload := data[payloadStart : len(data)-1]

	fields := []core.Field{
		newField(1, "协议头", 0, data[:5], "*HQ20", "VDF普通协议头"),
		newField(2, "回复属性", 5, data[5:6], replyText(reply), "1表示平台需要回复，0表示无需回复"),
		newField(3, "终端ID", 6, data[6:comma], terminalID, "终端SIM卡号码/设备标识"),
		newField(4, "分隔符", comma, data[comma:comma+1], ",", "终端ID与功能码分隔符"),
		newField(5, "功能类型", comma+1, data[comma+1:comma+2], functionTypeName(functionType), "A状态类、B定位类、G其他信息"),
		newField(6, "功能项关键字", comma+2, data[comma+2:comma+3], keywordName(messageID), "与功能类型组合形成消息ID"),
	}

	mainData, additionalData := splitPayload(payload, payloadStart)
	dataResult := map[string]interface{}{
		"header": map[string]interface{}{
			"replyRequired": reply == '1', "terminalId": terminalID,
			"functionType": string(functionType), "keyword": string(keyword),
		},
	}
	if len(mainData.value) > 0 {
		value := decodeMainData(messageID, string(mainData.value))
		fields = append(fields, newField(len(fields)+1, "指令数据", mainData.offset, mainData.value, value, mainDataDescription(messageID)))
		dataResult["messageData"] = value
	}

	items := make([]map[string]interface{}, 0, len(additionalData))
	for _, segment := range additionalData {
		name, value, decoded := decodeAdditional(segment.value)
		fields = append(fields, newField(len(fields)+1, name, segment.offset, segment.value, value, additionalDescription(segment.value)))
		items = append(items, decoded)
	}
	dataResult["additional"] = items
	fields = append(fields, newField(len(fields)+1, "协议尾", len(data)-1, data[len(data)-1:], "#", "VDF普通协议结束符"))

	return &core.ParseResult{
		Protocol: p.Name(), MessageID: messageID, MessageName: messageName(messageID),
		Length: len(data), Raw: hex.EncodeToString(data), Fields: fields, Data: dataResult,
	}, nil
}

type segment struct {
	offset int
	value  []byte
}

func splitPayload(payload []byte, base int) (segment, []segment) {
	first := len(payload)
	if index := strings.IndexByte(string(payload), '&'); index >= 0 {
		first = index
	}
	main := segment{offset: base, value: payload[:first]}
	items := make([]segment, 0)
	for position := first; position < len(payload); {
		next := strings.IndexByte(string(payload[position+1:]), '&')
		end := len(payload)
		if next >= 0 {
			end = position + 1 + next
		}
		items = append(items, segment{offset: base + position, value: payload[position:end]})
		position = end
	}
	return main, items
}

func newField(index int, name string, offset int, raw []byte, value, description string) core.Field {
	return core.Field{Index: index, Name: name, Offset: offset, Length: len(raw), Raw: hex.EncodeToString(raw), Value: value, Description: description}
}

func replyText(value byte) string {
	if value == '1' {
		return "需要平台回复"
	}
	if value == '0' {
		return "无需平台回复"
	}
	return fmt.Sprintf("未知回复属性（%c）", value)
}

func functionTypeName(value byte) string {
	switch value {
	case 'A':
		return "状态类（A）"
	case 'B':
		return "定位类（B）"
	case 'G':
		return "其他信息（G）"
	case 'R':
		return "报告类（R）"
	default:
		return fmt.Sprintf("功能类型（%c）", value)
	}
}

func messageName(id string) string {
	switch id {
	case "AB":
		return "设备登录"
	case "AW":
		return "远程升级回复"
	case "AX":
		return "设备参数回复"
	case "AY":
		return "设备上线提示"
	case "BT":
		return "蓝牙透传信息"
	case "BA":
		return "定时定位上报"
	case "GB":
		return "iBeacon扫描信息"
	case "GT":
		return "蓝牙Tag扫描信息"
	case "GD":
		return "蓝牙Tag移除信息"
	case "RP":
		return "报告数据"
	default:
		return "VDF消息（" + id + "）"
	}
}

func keywordName(id string) string { return messageName(id) }

func decodeMainData(id, value string) string {
	if id == "AB" && value == "1" {
		return "司机号：1（追踪模式或上电复位）"
	}
	if id == "AY" && value == "" {
		return "无指令数据"
	}
	return value
}

func mainDataDescription(id string) string {
	switch id {
	case "AB":
		return "登录司机号，追踪模式或上电复位时固定为1"
	case "BA":
		return "定位消息主体；详细信息位于&附加数据段"
	default:
		return "功能项对应的指令数据"
	}
}
