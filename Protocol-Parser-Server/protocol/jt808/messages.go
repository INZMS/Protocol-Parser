package jt808

import (
	"encoding/hex"
	"fmt"
	"strings"

	"protocol-parser-server/parser/core"
)

func parseBody(header *Header, body []byte, fields *[]core.Field) (map[string]interface{}, error) {
	base := 1 + header.Length
	switch header.MessageID {
	case 0x0001, 0x8001:
		return parseCommonResponse(body, base, header.MessageID, fields)
	case 0x0002, 0x0003, 0x8201:
		if len(body) != 0 {
			return nil, fmt.Errorf("%s消息体应为空，实际%d字节", messageName(header.MessageID), len(body))
		}
		return map[string]interface{}{"emptyBody": true}, nil
	case 0x0100:
		return parseRegistration(body, base, fields)
	case 0x8100:
		return parseRegistrationResponse(body, base, fields)
	case 0x0102:
		return parseAuthentication(body, base, fields), nil
	case 0x8103:
		return parseParameters(body, base, fields)
	case 0x8105:
		return parseTerminalControl(body, base, fields)
	case 0x0200:
		return parseLocationData(body, base, "", fields)
	case 0x0201:
		if err := requireLength(body, 2, "位置信息查询应答"); err != nil {
			return nil, err
		}
		*fields = append(*fields, newField(len(*fields)+1, "应答流水号", base, body[:2], fmt.Sprint(u16(body[:2])), "对应位置信息查询消息的流水号"))
		location, err := parseLocationData(body[2:], base+2, "", fields)
		if err != nil {
			return nil, err
		}
		location["responseSerial"] = u16(body[:2])
		return location, nil
	case 0x8202:
		return parseTrackingControl(body, base, fields)
	case 0x0704:
		return parseBatchLocation(body, base, fields)
	case 0x8300:
		return parseTextMessage(body, base, fields, false)
	case 0x6006:
		return parseTextMessage(body, base, fields, true)
	default:
		if len(body) > 0 {
			*fields = append(*fields, newField(len(*fields)+1, "消息体", base, body, hex.EncodeToString(body), "该消息ID尚未定义专用解析器"))
		}
		return map[string]interface{}{"bodyHex": hex.EncodeToString(body)}, nil
	}
}

func parseCommonResponse(body []byte, base int, messageID uint16, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 5, messageName(messageID)); err != nil {
		return nil, err
	}
	resultNames := map[byte]string{0: "成功/确认", 1: "失败", 2: "消息有误", 3: "不支持", 4: "报警处理确认"}
	resultText := resultNames[body[4]]
	if resultText == "" {
		resultText = fmt.Sprintf("未知结果%d", body[4])
	}
	*fields = append(*fields,
		newField(len(*fields)+1, "应答流水号", base, body[0:2], fmt.Sprint(u16(body[0:2])), "对应原消息的流水号"),
		newField(len(*fields)+2, "应答ID", base+2, body[2:4], fmt.Sprintf("0x%04X（%s）", u16(body[2:4]), messageName(u16(body[2:4]))), "对应原消息ID"),
		newField(len(*fields)+3, "应答结果", base+4, body[4:5], resultText, "应答处理结果"),
	)
	return map[string]interface{}{"responseSerial": u16(body[0:2]), "responseMessageId": fmt.Sprintf("%04x", u16(body[2:4])), "resultCode": body[4], "result": resultText}, nil
}

func parseRegistration(body []byte, base int, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 37, "终端注册"); err != nil {
		return nil, err
	}
	manufacturer := decodeGBK(body[4:9])
	model := decodeGBK(body[9:29])
	terminalID := decodeGBK(body[29:36])
	plate := decodeGBK(body[37:])
	*fields = append(*fields,
		newField(len(*fields)+1, "省域ID", base, body[0:2], fmt.Sprint(u16(body[0:2])), "GB/T 2260行政区划代码前两位"),
		newField(len(*fields)+2, "市县域ID", base+2, body[2:4], fmt.Sprint(u16(body[2:4])), "GB/T 2260行政区划代码后四位"),
		newField(len(*fields)+3, "制造商ID", base+4, body[4:9], manufacturer, "BYTE[5]"),
		newField(len(*fields)+4, "终端型号", base+9, body[9:29], model, "BYTE[20]，尾部0x00填充"),
		newField(len(*fields)+5, "终端ID", base+29, body[29:36], terminalID, "BYTE[7]，大写字母和数字"),
		newField(len(*fields)+6, "车牌颜色", base+36, body[36:37], fmt.Sprint(body[36]), "JT/T 415-2006，0表示未上牌"),
	)
	if len(body) > 37 {
		*fields = append(*fields, newField(len(*fields)+1, "车牌/VIN", base+37, body[37:], plate, "GBK；车牌颜色为0时表示VIN"))
	}
	return map[string]interface{}{"provinceId": u16(body[0:2]), "cityId": u16(body[2:4]), "manufacturerId": manufacturer, "terminalModel": model, "terminalId": terminalID, "plateColor": body[36], "plateNumber": plate}, nil
}

func parseRegistrationResponse(body []byte, base int, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 3, "终端注册应答"); err != nil {
		return nil, err
	}
	resultNames := map[byte]string{0: "成功", 1: "车辆已被注册", 2: "数据库中无该车辆", 3: "终端已被注册", 4: "数据库中无该车辆"}
	result := resultNames[body[2]]
	if result == "" {
		result = fmt.Sprintf("未知结果%d", body[2])
	}
	*fields = append(*fields,
		newField(len(*fields)+1, "应答流水号", base, body[:2], fmt.Sprint(u16(body[:2])), "对应终端注册消息流水号"),
		newField(len(*fields)+2, "注册结果", base+2, body[2:3], result, "0表示成功"),
	)
	authCode := ""
	if len(body) > 3 {
		authCode = decodeGBK(body[3:])
		*fields = append(*fields, newField(len(*fields)+1, "鉴权码", base+3, body[3:], authCode, "注册成功时返回"))
	}
	return map[string]interface{}{"responseSerial": u16(body[:2]), "resultCode": body[2], "result": result, "authenticationCode": authCode}, nil
}

func parseAuthentication(body []byte, base int, fields *[]core.Field) map[string]interface{} {
	value := decodeGBK(body)
	if len(body) > 0 {
		*fields = append(*fields, newField(len(*fields)+1, "鉴权码", base, body, value, "终端重连上报鉴权码"))
	}
	return map[string]interface{}{"authenticationCode": value}
}

func parseParameters(body []byte, base int, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 1, "设置终端参数"); err != nil {
		return nil, err
	}
	count := int(body[0])
	*fields = append(*fields, newField(len(*fields)+1, "参数总数", base, body[:1], fmt.Sprint(count), "后续参数项数量"))
	parameters := make([]map[string]interface{}, 0, count)
	offset := 1
	for index := 0; index < count; index++ {
		if len(body)-offset < 5 {
			return nil, fmt.Errorf("设置终端参数第%d项头部长度不足", index+1)
		}
		parameterID := u32(body[offset : offset+4])
		valueLength := int(body[offset+4])
		if len(body)-offset-5 < valueLength {
			return nil, fmt.Errorf("设置终端参数0x%08X声明长度%d，剩余%d字节", parameterID, valueLength, len(body)-offset-5)
		}
		value := body[offset+5 : offset+5+valueLength]
		text, decoded := parameterValue(parameterID, value)
		raw := body[offset : offset+5+valueLength]
		*fields = append(*fields, newField(len(*fields)+1, fmt.Sprintf("参数%d：0x%08X", index+1, parameterID), base+offset, raw, text, parameterDescription(parameterID)))
		parameters = append(parameters, map[string]interface{}{"id": fmt.Sprintf("%08x", parameterID), "length": valueLength, "value": decoded, "text": text})
		offset += 5 + valueLength
	}
	if offset != len(body) {
		return nil, fmt.Errorf("设置终端参数解析结束后仍有%d个未声明字节", len(body)-offset)
	}
	return map[string]interface{}{"parameterCount": count, "parameters": parameters}, nil
}

func parameterValue(id uint32, value []byte) (string, interface{}) {
	if id == 0x0010 || id == 0x0013 {
		text := decodeGBK(value)
		return text, text
	}
	if len(value) == 4 {
		number := u32(value)
		suffix := ""
		switch id {
		case 0x0001, 0x0027, 0x0029, 0x0056:
			suffix = " 秒"
		case 0x0018:
			suffix = "（TCP端口）"
		case 0x0055:
			suffix = " km/h"
		case 0x0080:
			return fmt.Sprintf("%.1f km", float64(number)/10), float64(number) / 10
		}
		return fmt.Sprintf("%d%s", number, suffix), number
	}
	return hex.EncodeToString(value), hex.EncodeToString(value)
}

func parameterDescription(id uint32) string {
	names := map[uint32]string{0x0001: "终端心跳发送间隔", 0x0010: "主服务器APN", 0x0013: "主服务器地址/IP/域名", 0x0018: "服务器TCP端口", 0x0027: "休眠时汇报间隔", 0x0029: "缺省时间汇报间隔", 0x0055: "最高速度", 0x0056: "超速持续时间", 0x0080: "车辆里程表读数"}
	if name, ok := names[id]; ok {
		return name
	}
	return "未定义参数"
}

func parseTerminalControl(body []byte, base int, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 1, "终端控制"); err != nil {
		return nil, err
	}
	commands := map[byte]string{0x04: "终端复位", 0x64: "断油电", 0x65: "通油电"}
	command := commands[body[0]]
	if command == "" {
		command = fmt.Sprintf("未知命令0x%02X", body[0])
	}
	*fields = append(*fields, newField(len(*fields)+1, "命令字", base, body[:1], command, "终端控制命令"))
	parameters := ""
	if len(body) > 1 {
		parameters = decodeGBK(body[1:])
		*fields = append(*fields, newField(len(*fields)+1, "命令参数", base+1, body[1:], parameters, "GBK编码，字段以半角分号分隔"))
	}
	return map[string]interface{}{"commandCode": body[0], "command": command, "parameters": parameters}, nil
}

func parseTrackingControl(body []byte, base int, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 2, "临时位置跟踪控制"); err != nil {
		return nil, err
	}
	interval := u16(body[:2])
	*fields = append(*fields, newField(len(*fields)+1, "时间间隔", base, body[:2], fmt.Sprintf("%d 秒", interval), "0表示停止跟踪"))
	duration := uint32(0)
	if interval > 0 {
		if err := requireLength(body, 6, "临时位置跟踪控制"); err != nil {
			return nil, err
		}
		duration = u32(body[2:6])
		*fields = append(*fields, newField(len(*fields)+1, "位置跟踪有效期", base+2, body[2:6], fmt.Sprintf("%d 秒", duration), "跟踪控制有效期"))
	}
	return map[string]interface{}{"intervalSeconds": interval, "durationSeconds": duration, "stopped": interval == 0}, nil
}

func parseTextMessage(body []byte, base int, fields *[]core.Field, upstream bool) (map[string]interface{}, error) {
	if err := requireLength(body, 1, messageName(map[bool]uint16{true: 0x6006, false: 0x8300}[upstream])); err != nil {
		return nil, err
	}
	if upstream {
		encodingName := map[byte]string{0: "GB2312/GBK", 1: "Unicode UTF-16BE"}[body[0]]
		if encodingName == "" {
			encodingName = fmt.Sprintf("未知编码%d", body[0])
		}
		text := decodeGBK(body[1:])
		if body[0] == 1 {
			text = decodeUTF16BE(body[1:])
		}
		*fields = append(*fields,
			newField(len(*fields)+1, "文本编码方式", base, body[:1], encodingName, "0为GB2312，1为Unicode"),
			newField(len(*fields)+2, "文本消息", base+1, body[1:], text, encodingName),
		)
		return map[string]interface{}{"encodingCode": body[0], "encoding": encodingName, "text": text}, nil
	}
	emergency := body[0]&1 != 0
	text := decodeGBK(body[1:])
	*fields = append(*fields,
		newField(len(*fields)+1, "文本标志", base, body[:1], boolText(emergency, "紧急文本", "普通文本"), "bit0为紧急标志"),
		newField(len(*fields)+2, "文本信息", base+1, body[1:], text, "GBK编码，最长1024字节"),
	)
	return map[string]interface{}{"emergency": emergency, "text": text}, nil
}

func formatLabels(values []string) string {
	if len(values) == 0 {
		return "无"
	}
	return strings.Join(values, "、")
}
