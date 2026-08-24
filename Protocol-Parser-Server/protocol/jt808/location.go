package jt808

import (
	"encoding/hex"
	"fmt"
	"strings"

	"protocol-parser-server/parser/core"
)

func parseLocationData(body []byte, base int, prefix string, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 28, prefix+"位置信息"); err != nil {
		return nil, err
	}
	alarm := u32(body[0:4])
	status := u32(body[4:8])
	latitude := float64(u32(body[8:12])) / 1_000_000
	longitude := float64(u32(body[12:16])) / 1_000_000
	if status&(1<<2) != 0 {
		latitude = -latitude
	}
	if status&(1<<3) != 0 {
		longitude = -longitude
	}
	altitude := u16(body[16:18])
	speed := float64(u16(body[18:20])) / 10
	direction := u16(body[20:22])
	timeText, err := decodeTime(body[22:28])
	if err != nil {
		return nil, err
	}
	name := func(value string) string {
		if prefix == "" {
			return value
		}
		return prefix + value
	}
	*fields = append(*fields,
		newField(len(*fields)+1, name("报警标志"), base, body[0:4], alarmText(alarm), fmt.Sprintf("状态字0x%08X，按位解析", alarm)),
		newField(len(*fields)+2, name("车辆状态"), base+4, body[4:8], statusText(status), fmt.Sprintf("状态字0x%08X，按位解析", status)),
		newField(len(*fields)+3, name("纬度"), base+8, body[8:12], fmt.Sprintf("%.6f", latitude), "单位：度；DWORD值除以10^6，状态bit2确定南北纬"),
		newField(len(*fields)+4, name("经度"), base+12, body[12:16], fmt.Sprintf("%.6f", longitude), "单位：度；DWORD值除以10^6，状态bit3确定东西经"),
		newField(len(*fields)+5, name("高程"), base+16, body[16:18], fmt.Sprintf("%d m", altitude), "海拔高度"),
		newField(len(*fields)+6, name("速度"), base+18, body[18:20], fmt.Sprintf("%.1f km/h", speed), "WORD，单位0.1km/h"),
		newField(len(*fields)+7, name("方向"), base+20, body[20:22], fmt.Sprintf("%d°", direction), "0-359，正北为0，顺时针"),
		newField(len(*fields)+8, name("定位时间"), base+22, body[22:28], timeText, "BCD[6]，YY-MM-DD-hh-mm-ss，GMT+8"),
	)

	additional, err := parseLocationAdditional(body[28:], base+28, prefix, fields)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"alarm":  map[string]interface{}{"raw": fmt.Sprintf("%08X", alarm), "active": alarmLabels(alarm)},
		"status": statusData(status), "latitude": latitude, "longitude": longitude,
		"altitudeMeters": altitude, "speedKmh": speed, "directionDegrees": direction,
		"locationTime": timeText, "additional": additional,
	}, nil
}

func alarmLabels(value uint32) []string {
	definitions := []struct {
		bit  uint
		name string
	}{{0, "SOS"}, {1, "超速报警"}, {7, "主电源欠压"}, {8, "主电源掉电"}, {15, "震动告警"}, {30, "侧翻预警"}}
	labels := make([]string, 0)
	for _, definition := range definitions {
		if value&(1<<definition.bit) != 0 {
			labels = append(labels, definition.name)
		}
	}
	return labels
}

func alarmText(value uint32) string {
	labels := alarmLabels(value)
	if len(labels) == 0 {
		return "无报警"
	}
	return formatLabels(labels)
}

func statusData(value uint32) map[string]interface{} {
	return map[string]interface{}{
		"raw": fmt.Sprintf("%08X", value), "accOn": value&1 != 0, "positioned": value&(1<<1) != 0,
		"latitudeHemisphere":  boolText(value&(1<<2) != 0, "南纬", "北纬"),
		"longitudeHemisphere": boolText(value&(1<<3) != 0, "西经", "东经"),
		"armed":               value&(1<<6) != 0, "fuelCut": value&(1<<10) != 0,
		"mainPowerDisconnected": value&(1<<11) != 0, "alarmShielded": value&(1<<17) != 0,
		"removalAlarm": value&(1<<30) != 0,
	}
}

func statusText(value uint32) string {
	return strings.Join([]string{
		boolText(value&1 != 0, "ACC开", "ACC关"),
		boolText(value&(1<<1) != 0, "定位有效", "未定位"),
		boolText(value&(1<<2) != 0, "南纬", "北纬"),
		boolText(value&(1<<3) != 0, "西经", "东经"),
		boolText(value&(1<<6) != 0, "已设防", "已撤防"),
		boolText(value&(1<<10) != 0, "油路断开", "油路正常"),
		boolText(value&(1<<11) != 0, "主电源断开", "主电源接通"),
	}, "｜")
}

func parseLocationAdditional(data []byte, base int, prefix string, fields *[]core.Field) ([]map[string]interface{}, error) {
	items := make([]map[string]interface{}, 0)
	for offset := 0; offset < len(data); {
		if len(data)-offset < 2 {
			return nil, fmt.Errorf("位置附加信息项头部长度不足，偏移%d", offset)
		}
		id := data[offset]
		length := int(data[offset+1])
		if len(data)-offset-2 < length {
			return nil, fmt.Errorf("位置附加信息0x%02X声明长度%d，剩余%d字节", id, length, len(data)-offset-2)
		}
		value := data[offset+2 : offset+2+length]
		name, text, decoded := decodeLocationAdditional(id, value)
		fieldName := "附加信息：" + name
		if prefix != "" {
			fieldName = prefix + fieldName
		}
		*fields = append(*fields, newField(len(*fields)+1, fieldName, base+offset, data[offset:offset+2+length], text, fmt.Sprintf("附加信息ID 0x%02X，值长度%d字节", id, length)))
		items = append(items, map[string]interface{}{"id": fmt.Sprintf("%02x", id), "name": name, "length": length, "value": decoded, "text": text})
		offset += 2 + length
	}
	return items, nil
}

func decodeLocationAdditional(id byte, value []byte) (string, string, interface{}) {
	switch id {
	case 0x01:
		if len(value) == 4 {
			mileage := float64(u32(value)) / 10
			return "里程", fmt.Sprintf("%.1f km", mileage), mileage
		}
	case 0x30:
		if len(value) == 1 {
			return "通讯信号强度", fmt.Sprintf("%d", value[0]), value[0]
		}
	case 0x31:
		if len(value) == 1 {
			return "有效定位卫星数", fmt.Sprintf("%d颗", value[0]), value[0]
		}
	case 0x54:
		return decodeWiFi(value)
	case 0x5d:
		return decodeBaseStations(value)
	case 0x61:
		if len(value) == 2 {
			voltage := float64(u16(value)) / 100
			return "主电源电压", fmt.Sprintf("%.2f V", voltage), voltage
		}
	case 0xf1:
		iccid := string(trimZero(value))
		return "SIM ICCID", iccid, iccid
	}
	encoded := hex.EncodeToString(value)
	return fmt.Sprintf("未知附加0x%02X", id), encoded, encoded
}

func decodeWiFi(value []byte) (string, string, interface{}) {
	if len(value) < 1 {
		return "WiFi热点", "数据为空", []interface{}{}
	}
	count := int(value[0])
	if len(value) != 1+count*7 {
		encoded := hex.EncodeToString(value)
		return "WiFi热点", fmt.Sprintf("长度不匹配：%s", encoded), encoded
	}
	hotspots := make([]map[string]interface{}, 0, count)
	parts := []string{fmt.Sprintf("%d个WiFi热点", count)}
	for index := 0; index < count; index++ {
		item := value[1+index*7 : 1+(index+1)*7]
		mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", item[0], item[1], item[2], item[3], item[4], item[5])
		signal := int(int8(item[6]))
		hotspots = append(hotspots, map[string]interface{}{"mac": mac, "signalDbm": signal})
		parts = append(parts, fmt.Sprintf("#%d %s（%d dBm）", index+1, mac, signal))
	}
	return "WiFi热点", strings.Join(parts, "；"), hotspots
}

func decodeBaseStations(value []byte) (string, string, interface{}) {
	if len(value) < 1 {
		return "基站信息", "数据为空", []interface{}{}
	}
	count := int(value[0])
	if len(value) != 1+count*10 {
		encoded := hex.EncodeToString(value)
		return "基站信息", fmt.Sprintf("长度不匹配：%s", encoded), encoded
	}
	stations := make([]map[string]interface{}, 0, count)
	parts := []string{fmt.Sprintf("%d个基站", count)}
	for index := 0; index < count; index++ {
		item := value[1+index*10 : 1+(index+1)*10]
		station := map[string]interface{}{"mcc": u16(item[0:2]), "mnc": item[2], "lac": u16(item[3:5]), "cellId": u32(item[5:9]), "signalDbm": int(int8(item[9]))}
		stations = append(stations, station)
		parts = append(parts, fmt.Sprintf("#%d MCC=%d，MNC=%d，LAC=%d，CellID=%d，信号=%d dBm", index+1, station["mcc"], station["mnc"], station["lac"], station["cellId"], station["signalDbm"]))
	}
	return "基站信息", strings.Join(parts, "；"), stations
}

func parseBatchLocation(body []byte, base int, fields *[]core.Field) (map[string]interface{}, error) {
	if err := requireLength(body, 3, "盲区定位数据批量上传"); err != nil {
		return nil, err
	}
	count := int(u16(body[:2]))
	reportType := body[2]
	reportTypeText := map[byte]string{0: "正常位置批量汇报", 1: "盲区补报"}[reportType]
	if reportTypeText == "" {
		reportTypeText = fmt.Sprintf("未知类型%d", reportType)
	}
	*fields = append(*fields,
		newField(len(*fields)+1, "位置数据项个数", base, body[:2], fmt.Sprint(count), "WORD，大端字节序"),
		newField(len(*fields)+2, "位置数据类型", base+2, body[2:3], reportTypeText, "0正常批量汇报，1盲区补报"),
	)
	locations := make([]map[string]interface{}, 0, count)
	offset := 3
	for index := 0; index < count; index++ {
		if len(body)-offset < 2 {
			return nil, fmt.Errorf("批量位置第%d项缺少长度字段", index+1)
		}
		length := int(u16(body[offset : offset+2]))
		if len(body)-offset-2 < length {
			return nil, fmt.Errorf("批量位置第%d项声明长度%d，剩余%d字节", index+1, length, len(body)-offset-2)
		}
		*fields = append(*fields, newField(len(*fields)+1, fmt.Sprintf("位置项%d长度", index+1), base+offset, body[offset:offset+2], fmt.Sprintf("%d Bytes", length), "后续位置汇报数据体长度"))
		location, err := parseLocationData(body[offset+2:offset+2+length], base+offset+2, fmt.Sprintf("位置项%d：", index+1), fields)
		if err != nil {
			return nil, err
		}
		locations = append(locations, location)
		offset += 2 + length
	}
	if offset != len(body) {
		return nil, fmt.Errorf("批量位置解析结束后仍有%d个未声明字节", len(body)-offset)
	}
	return map[string]interface{}{"itemCount": count, "reportTypeCode": reportType, "reportType": reportTypeText, "locations": locations}, nil
}
