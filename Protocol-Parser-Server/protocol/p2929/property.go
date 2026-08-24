package p2929

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ReportProperty struct {
	MessageType      string                 `json:"messageType"`
	MessageID        string                 `json:"messageId"`
	DeviceNo         string                 `json:"deviceNo"`
	ProductNo        string                 `json:"productNo,omitempty"`
	Timestamp        int64                  `json:"timestamp"`
	ReceiptTimestamp int64                  `json:"receiptTimestamp,omitempty"`
	PropertiesMap    map[string]interface{} `json:"propertiesMap"`
	AsyncSave        bool                   `json:"asyncSave"`
	FlowType         string                 `json:"flowType"`
}

func ParseReportProperty(header *Header, data []byte) (*ReportProperty, []TLV, error) {
	if len(data) < headerLength+locationBaseLength+trailerLength {
		return nil, nil, fmt.Errorf("0x80报文长度不足: 至少需要%d字节, 实际%d字节", headerLength+locationBaseLength+trailerLength, len(data))
	}
	body := data[headerLength : len(data)-trailerLength]
	location, err := ParseLocationBase(body)
	if err != nil {
		return nil, nil, err
	}
	result := &ReportProperty{MessageType: "reportProperty", MessageID: fmt.Sprintf("%02x", header.Cmd), DeviceNo: decodePseudoIP(header.IP), Timestamp: location.DeviceTime, PropertiesMap: map[string]interface{}{}, AsyncSave: true, FlowType: "upstream"}
	result.PropertiesMap["location"] = map[string]interface{}{"lat": location.Lat, "lng": location.Lng}
	result.PropertiesMap["speed"] = location.Speed
	result.PropertiesMap["direction"] = location.Direction
	result.PropertiesMap["locationStatus"] = map[string]interface{}{"valid": location.HasValidLocation, "gpsAntenna": location.GPSAntenna, "power": location.PowerStatus, "raw": fmt.Sprintf("%02X", location.Status)}
	result.PropertiesMap["vehicleStatus"] = map[string]interface{}{"raw": fmt.Sprintf("%08X", location.VehicleStatus), "needAck": location.NeedAck, "transport": location.Transport, "signalStrength": location.SignalStrength}
	result.PropertiesMap["centerCommand"] = fmt.Sprintf("0x%02X", location.CenterCommand)
	result.PropertiesMap["deviceTime"] = location.DeviceTime
	result.PropertiesMap["deviceTimeText"] = timeText(location.DeviceTime)
	result.PropertiesMap["timeZone"] = "Asia/Shanghai"
	extension := body[locationBaseLength:]
	result.PropertiesMap["rawExtension"] = hex.EncodeToString(extension)
	tlvs, err := ParseTLV(extension)
	if err != nil {
		return nil, nil, err
	}
	for _, item := range tlvs {
		parseExtension(result, item)
	}
	return result, tlvs, nil
}

func parseExtension(result *ReportProperty, item TLV) {
	raw := hex.EncodeToString(item.Value)
	switch item.Type {
	case 0x0024:
		result.PropertiesMap["singleBaseStation"] = string(item.Value)
	case 0x0004:
		if len(item.Value) == 3 {
			result.PropertiesMap["voltage"] = float64(bcdInt(item.Value)) / 100000
		} else {
			result.PropertiesMap["voltageRaw"] = raw
		}
	case 0x0008:
		if len(item.Value) >= 2 {
			count := int(item.Value[0])<<8 | int(item.Value[1])
			result.PropertiesMap["battery"] = map[string]interface{}{"count": count, "total": 1500, "percent": float64(count) / 15}
		}
	case 0x00A3:
		result.PropertiesMap["softwareVersion"] = strings.TrimRight(string(item.Value), "\x00;")
	case 0x00A5:
		result.ProductNo = raw
		result.PropertiesMap["terminalModel"] = raw
	case 0x0089:
		result.PropertiesMap["alarmStatus"] = map[string]interface{}{"raw": raw, "moving": bit(item.Value, 9) == 0, "sos": bit(item.Value, 10) == 0, "removed": bit(item.Value, 12) == 0}
	case 0x00A9:
		result.PropertiesMap["baseStations"] = parseBaseStations(item.Value)
	case 0x00B9:
		result.PropertiesMap["wifi"] = parseWifi(item.Value)
	case 0x00C5:
		result.PropertiesMap["extendedLocationStatus"] = parseExtendedLocation(item.Value)
	case 0x00FB:
		result.PropertiesMap["iccid"] = strings.TrimRight(string(item.Value), "\x00")
	case 0x00AE:
		result.PropertiesMap["workMode"] = parseWorkMode(item.Value)
	case 0xF000:
		result.PropertiesMap["reportMode"] = parseReportMode(item.Value)
	case 0xF001:
		result.PropertiesMap["nextReport"] = parseNextReport(item.Value)
	case 0x0030:
		result.PropertiesMap["signalStrength"] = first(item.Value)
	case 0x0031:
		result.PropertiesMap["satellites"] = first(item.Value)
	default:
		result.PropertiesMap[fmt.Sprintf("unknown_%04x", item.Type)] = raw
	}
}

func parseWifi(data []byte) map[string]interface{} {
	if len(data) == 0 {
		return map[string]interface{}{"count": 0, "hotspots": []map[string]interface{}{}}
	}
	parts := strings.Split(string(data[1:]), ",")
	count := int(data[0])
	hotspots := make([]map[string]interface{}, 0, count)
	for i := 0; i+1 < len(parts) && len(hotspots) < count; i += 2 {
		signal, err := strconv.Atoi(strings.TrimSpace(parts[i+1]))
		if err != nil {
			continue
		}
		hotspots = append(hotspots, map[string]interface{}{
			"mac":    strings.TrimSpace(parts[i]),
			"signal": signal,
		})
	}
	return map[string]interface{}{"count": count, "hotspots": hotspots}
}
func parseBaseStations(data []byte) map[string]interface{} {
	result := map[string]interface{}{"raw": hex.EncodeToString(data)}
	if len(data) < 4 {
		return result
	}
	result["country"] = int(data[0])<<8 | int(data[1])
	result["operator"] = int(data[2])
	result["count"] = int(data[3])
	stations := []map[string]int{}
	for i := 4; i+5 <= len(data) && len(stations) < result["count"].(int); i += 5 {
		stations = append(stations, map[string]int{"area": int(data[i])<<8 | int(data[i+1]), "tower": int(data[i+2])<<8 | int(data[i+3]), "signal": int(data[i+4])})
	}
	result["stations"] = stations
	return result
}
func parseExtendedLocation(data []byte) map[string]interface{} {
	value := uint32(0)
	for _, v := range data {
		value = value<<8 | uint32(v)
	}
	mode := "unknown"
	if value&0x08 != 0 {
		mode = "gps"
	} else if value&0x10 != 0 {
		mode = "wifi"
	} else {
		mode = "none"
	}
	return map[string]interface{}{"raw": fmt.Sprintf("%08X", value), "mode": mode}
}
func parseNextReport(data []byte) interface{} {
	if len(data) < 7 {
		return hex.EncodeToString(data)
	}
	timestamp, err := parse2929Time(data[:6])
	if err != nil {
		return hex.EncodeToString(data)
	}
	environmentCode := data[6]
	environment := map[byte]string{0: "全天空", 1: "半天空", 2: "地下室"}[environmentCode]
	if environment == "" {
		environment = fmt.Sprintf("未知环境（%d）", environmentCode)
	}
	return map[string]interface{}{
		"timestamp":       timestamp,
		"timeText":        timeText(timestamp),
		"timeZone":        "Asia/Shanghai",
		"environment":     environment,
		"environmentCode": int(environmentCode),
	}
}

func workModeLabel(mode int) string {
	labels := map[int]string{1: "闹钟模式", 3: "追踪模式", 4: "星期模式", 6: "月模式"}
	if label, ok := labels[mode]; ok {
		return label
	}
	return fmt.Sprintf("未知工作模式（%d）", mode)
}

func reportModeLabel(mode int) string {
	labels := map[int]string{0: "闹钟模式", 1: "定时回传", 2: "星期模式"}
	if label, ok := labels[mode]; ok {
		return label
	}
	return fmt.Sprintf("未知上报模式（%d）", mode)
}

func parseWorkMode(data []byte) map[string]interface{} {
	mode := first(data)
	result := map[string]interface{}{
		"mode":  mode,
		"label": workModeLabel(mode),
		"raw":   hex.EncodeToString(data),
	}
	parameters := data[min(1, len(data)):]
	switch mode {
	case 0x01:
		result["times"] = parseBCDClocks(parameters, 4)
	case 0x03:
		if len(parameters) >= 2 {
			result["intervalMinutes"] = int(parameters[0])<<8 | int(parameters[1])
		}
	case 0x04, 0x06:
		if len(parameters) >= 2 {
			result["enabled"] = parameters[0] == 1
			count := min(int(parameters[1]), len(parameters)-2)
			values := make([]int, count)
			for i := 0; i < count; i++ {
				values[i] = int(parameters[2+i])
			}
			if mode == 0x04 {
				result["weekdays"] = weekdayNames(values)
			} else {
				result["dates"] = values
			}
			if clock, ok := parseBCDClock(parameters[2+count:]); ok {
				result["time"] = clock
			}
		}
	}
	result["summary"] = modeSummary(result, true)
	return result
}

func parseReportMode(data []byte) map[string]interface{} {
	mode := first(data)
	result := map[string]interface{}{
		"mode":  mode,
		"label": reportModeLabel(mode),
		"raw":   hex.EncodeToString(data),
	}
	parameters := data[min(1, len(data)):]
	switch mode {
	case 0x00:
		result["times"] = parseBCDClocks(parameters, 4)
	case 0x01:
		if len(parameters) >= 2 {
			result["intervalMinutes"] = int(parameters[0])<<8 | int(parameters[1])
		}
	case 0x02:
		if len(parameters) >= 3 {
			result["weekdays"] = weekdayNamesFromBits(parameters[0])
			if clock, ok := parseBCDClock(parameters[1:]); ok {
				result["time"] = clock
			}
		}
	}
	result["summary"] = modeSummary(result, false)
	return result
}

func modeSummary(value map[string]interface{}, workMode bool) string {
	parts := []string{fmt.Sprintf("%v（模式值：%v）", value["label"], value["mode"])}
	if times, ok := value["times"].([]string); ok && len(times) > 0 {
		label := "上报时间"
		if workMode {
			label = "唤醒时间"
		}
		parts = append(parts, fmt.Sprintf("%s：%s", label, strings.Join(times, "、")))
	}
	if interval, ok := value["intervalMinutes"].(int); ok {
		parts = append(parts, fmt.Sprintf("回传间隔：%d分钟", interval))
	}
	if enabled, ok := value["enabled"].(bool); ok {
		if enabled {
			parts = append(parts, "已开启")
		} else {
			parts = append(parts, "未开启")
		}
	}
	if weekdays, ok := value["weekdays"].([]string); ok && len(weekdays) > 0 {
		parts = append(parts, "日期："+strings.Join(weekdays, "、"))
	}
	if dates, ok := value["dates"].([]int); ok && len(dates) > 0 {
		items := make([]string, len(dates))
		for i, day := range dates {
			items[i] = fmt.Sprintf("%d日", day)
		}
		parts = append(parts, "日期："+strings.Join(items, "、"))
	}
	if clock, ok := value["time"].(string); ok {
		label := "上报时间"
		if workMode {
			label = "唤醒时间"
		}
		parts = append(parts, label+"："+clock)
	}
	return strings.Join(parts, "｜")
}

func parseBCDClocks(data []byte, limit int) []string {
	result := make([]string, 0, min(len(data)/2, limit))
	for i := 0; i+2 <= len(data) && len(result) < limit; i += 2 {
		if clock, ok := parseBCDClock(data[i : i+2]); ok {
			result = append(result, clock)
		}
	}
	return result
}

func parseBCDClock(data []byte) (string, bool) {
	if len(data) < 2 || data[0]>>4 > 9 || data[0]&0x0F > 9 || data[1]>>4 > 9 || data[1]&0x0F > 9 {
		return "", false
	}
	hour, minute := bcd(data[0]), bcd(data[1])
	if hour > 23 || minute > 59 {
		return "", false
	}
	return fmt.Sprintf("%02d:%02d", hour, minute), true
}

func weekdayNames(values []int) []string {
	labels := map[int]string{1: "星期一", 2: "星期二", 3: "星期三", 4: "星期四", 5: "星期五", 6: "星期六", 7: "星期日"}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if label, ok := labels[value]; ok {
			result = append(result, label)
		}
	}
	return result
}

func weekdayNamesFromBits(bits byte) []string {
	values := make([]int, 0, 7)
	for index := 0; index < 7; index++ {
		if bits&(1<<index) != 0 {
			values = append(values, index+1)
		}
	}
	return weekdayNames(values)
}
func bit(data []byte, index uint) byte {
	value := uint64(0)
	for _, v := range data {
		value = value<<8 | uint64(v)
	}
	return byte(value >> index & 1)
}
func first(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return int(data[0])
}

func decodePseudoIP(data []byte) string {
	if len(data) != 4 {
		return ""
	}
	marker := 0
	groups := make([]string, 4)
	for i, v := range data {
		if v&0x80 != 0 {
			marker |= 1 << (3 - i)
		}
		groups[i] = fmt.Sprintf("%02d", v&0x7F)
	}
	return fmt.Sprintf("1%02d%s", marker+30, strings.Join(groups, ""))
}
func parse2929Time(data []byte) (int64, error) {
	if len(data) != 6 {
		return 0, fmt.Errorf("BCD时间需要6字节, 实际%d字节", len(data))
	}
	for _, v := range data {
		if v>>4 > 9 || v&15 > 9 {
			return 0, fmt.Errorf("无效BCD值%02X", v)
		}
	}
	year := 2000 + bcd(data[0])
	month := time.Month(bcd(data[1]))
	day := bcd(data[2])
	hour := bcd(data[3])
	minute := bcd(data[4])
	second := bcd(data[5])
	value := time.Date(year, month, day, hour, minute, second, 0, beijingLocation)
	if value.Year() != year || value.Month() != month || value.Day() != day || value.Hour() != hour || value.Minute() != minute || value.Second() != second {
		return 0, fmt.Errorf("无效日期时间")
	}
	return value.UnixMilli(), nil
}
func bcd(v byte) int { return int(v>>4)*10 + int(v&15) }
