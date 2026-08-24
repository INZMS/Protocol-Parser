package p2929

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"protocol-parser-server/parser/core"
)

var beijingLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

func ParseLocation(p *Protocol2929, header *Header, data []byte) (*core.ParseResult, error) {
	property, tlvs, err := ParseReportProperty(header, data)
	if err != nil {
		return nil, err
	}
	body := data[headerLength : len(data)-trailerLength]
	location := property.PropertiesMap["location"].(map[string]interface{})
	status := property.PropertiesMap["locationStatus"].(map[string]interface{})
	vehicle := property.PropertiesMap["vehicleStatus"].(map[string]interface{})
	fields := headerFields(header, data)
	fields = append(fields,
		newField(5, "定位时间", 9, body[0:6], timeText(property.Timestamp), "北京时间（UTC+8），YYMMDDHHmmss压缩BCD"),
		newField(6, "纬度", 15, body[6:10], fmt.Sprint(location["lat"]), "DDMM.mmm，最高位为南纬符号"),
		newField(7, "经度", 19, body[10:14], fmt.Sprint(location["lng"]), "DDDMM.mmm，最高位为西经符号"),
		newField(8, "速度", 23, body[14:16], fmt.Sprintf("%v km/h", property.PropertiesMap["speed"]), "压缩BCD"),
		newField(9, "方向", 25, body[16:18], fmt.Sprintf("%v°", property.PropertiesMap["direction"]), "正北0度，顺时针"),
		newField(10, "定位/天线/电源状态", 27, body[18:19], locationStatusText(status), fmt.Sprintf("状态字0x%02X，按位解析", body[18])),
		newField(11, "保留字段 LLL", 28, body[19:22], "协议保留（3字节）", "保留字段，不参与业务解析"),
		newField(12, "车辆状态 ABCD", 31, body[22:26], vehicleStatusText(vehicle), fmt.Sprintf("状态字0x%08X，按位解析", locationVehicleRaw(vehicle))),
		newField(13, "保留字段 WWERTYU", 35, body[26:33], "协议保留（7字节）", "保留字段，不参与业务解析"),
		newField(14, "中心命令", 42, body[33:34], centerCommandText(body[33]), "中心下发的主命令"),
	)
	for _, item := range tlvs {
		start := headerLength + locationBaseLength + item.Offset
		raw := data[start : start+int(item.Length)+2]
		fields = append(fields, newField(len(fields)+1, extensionName(item.Type), start, raw, extensionValue(property, item), fmt.Sprintf("扩展指令0x%04X，长度字段包含指令字", item.Type)))
	}
	fields = append(fields, trailerFields(len(fields)+1, data)...)
	return &core.ParseResult{Protocol: p.Name(), MessageID: hex.EncodeToString([]byte{header.Cmd}), MessageName: MessageName(header.Cmd), Length: len(data), Data: property, Raw: hex.EncodeToString(data), Fields: fields}, nil
}

func locationStatusText(status map[string]interface{}) string {
	locationText := "未定位"
	if valid, ok := status["valid"].(bool); ok && valid {
		locationText = "定位有效"
	}
	return fmt.Sprintf("%s｜天线%s｜电源%s", locationText, status["gpsAntenna"], status["power"])
}

func vehicleStatusText(vehicle map[string]interface{}) string {
	ack := "无需应答"
	if needAck, ok := vehicle["needAck"].(bool); ok && needAck {
		ack = "需要应答"
	}
	return fmt.Sprintf("%v传输｜信号强度%v/31｜%s", vehicle["transport"], vehicle["signalStrength"], ack)
}

func locationVehicleRaw(vehicle map[string]interface{}) uint32 {
	raw, _ := vehicle["raw"].(string)
	var value uint32
	fmt.Sscanf(raw, "%08X", &value)
	return value
}

func centerCommandText(command byte) string {
	if command == 0 {
		return "无中心命令"
	}
	return fmt.Sprintf("中心命令0x%02X", command)
}

func extensionName(command uint16) string {
	names := map[uint16]string{0x0024: "单基站", 0x0004: "AD电压", 0x0008: "电池电量", 0x00A3: "软件版本", 0x00A5: "终端型号", 0x0089: "扩展报警状态", 0x00A9: "多基站", 0x00B9: "WiFi热点", 0x00C5: "扩展定位状态", 0x00FB: "SIM ICCID", 0x00AE: "当前工作模式", 0xF000: "预设上报模式", 0xF001: "下次上报时间", 0x0030: "通讯信号强度", 0x0031: "GNSS卫星数"}
	if name, ok := names[command]; ok {
		return name
	}
	return fmt.Sprintf("未知扩展0x%04X", command)
}
func extensionValue(property *ReportProperty, item TLV) string {
	keys := map[uint16]string{0x0024: "singleBaseStation", 0x0004: "voltage", 0x0008: "battery", 0x00A3: "softwareVersion", 0x00A5: "terminalModel", 0x0089: "alarmStatus", 0x00A9: "baseStations", 0x00B9: "wifi", 0x00C5: "extendedLocationStatus", 0x00FB: "iccid", 0x00AE: "workMode", 0xF000: "reportMode", 0xF001: "nextReport", 0x0030: "signalStrength", 0x0031: "satellites"}
	if key, ok := keys[item.Type]; ok {
		value := property.PropertiesMap[key]
		switch item.Type {
		case 0x0008:
			if battery, ok := value.(map[string]interface{}); ok {
				return fmt.Sprintf("%.2f%%（%v/%v）", battery["percent"], battery["count"], battery["total"])
			}
		case 0x0089:
			if alarm, ok := value.(map[string]interface{}); ok {
				return fmt.Sprintf("运动=%v，SOS=%v，拆除=%v", alarm["moving"], alarm["sos"], alarm["removed"])
			}
		case 0x00A9:
			if stations, ok := value.(map[string]interface{}); ok {
				return baseStationsText(stations)
			}
		case 0x00B9:
			if wifi, ok := value.(map[string]interface{}); ok {
				return wifiText(wifi)
			}
		case 0x00C5:
			if status, ok := value.(map[string]interface{}); ok {
				labels := map[interface{}]string{"gps": "GPS定位", "wifi": "WiFi定位", "none": "未定位"}
				if label, exists := labels[status["mode"]]; exists {
					return label
				}
			}
		case 0x00AE:
			if mode, ok := value.(map[string]interface{}); ok {
				return fmt.Sprint(mode["summary"])
			}
		case 0xF000:
			if mode, ok := value.(map[string]interface{}); ok {
				return fmt.Sprint(mode["summary"])
			}
		case 0xF001:
			if next, ok := value.(map[string]interface{}); ok {
				return fmt.Sprintf("时间：%v；定位环境：%v（环境值：%v）", next["timeText"], next["environment"], next["environmentCode"])
			}
		}
		return fmt.Sprint(value)
	}
	return hex.EncodeToString(item.Value)
}

func baseStationsText(value map[string]interface{}) string {
	parts := []string{fmt.Sprintf("%v个基站（MCC=%v，MNC=%v）", value["count"], value["country"], value["operator"])}
	if stations, ok := value["stations"].([]map[string]int); ok {
		for index, station := range stations {
			parts = append(parts, fmt.Sprintf("#%d LAC=%d，CellID=%d，信号=%d", index+1, station["area"], station["tower"], station["signal"]))
		}
	}
	return strings.Join(parts, "；")
}

func wifiText(value map[string]interface{}) string {
	parts := []string{fmt.Sprintf("%v个WiFi热点", value["count"])}
	if hotspots, ok := value["hotspots"].([]map[string]interface{}); ok {
		for index, hotspot := range hotspots {
			parts = append(parts, fmt.Sprintf("#%d %v（%v dBm）", index+1, hotspot["mac"], hotspot["signal"]))
		}
	}
	return strings.Join(parts, "；")
}
func timeText(timestamp int64) string {
	return time.UnixMilli(timestamp).In(beijingLocation).Format("2006-01-02 15:04:05") + "（北京时间）"
}
