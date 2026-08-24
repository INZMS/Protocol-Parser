package vdf

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var beijingLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

func decodeAdditional(raw []byte) (string, string, map[string]interface{}) {
	text := string(raw)
	if len(text) < 2 || text[0] != '&' {
		return "附加信息", text, map[string]interface{}{"raw": text}
	}
	typeCode := text[1]
	value := text[2:]
	name := additionalName(typeCode)
	decoded := map[string]interface{}{"type": string(typeCode), "raw": value}
	display := value

	switch typeCode {
	case 'A':
		if result, label, ok := decodeGPS(value); ok {
			decoded = result
			decoded["type"] = "A"
			display = label
		}
	case 'B':
		if status, alarms, label, ok := decodeStatus(value); ok {
			decoded["status"], decoded["alarms"], display = status, alarms, label
		}
	case 'F':
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			knots := number / 10
			kmh := knots * 1.852
			decoded["knots"], decoded["speedKmh"] = knots, kmh
			display = fmt.Sprintf("%.1f km/h（%.1f节）", kmh, knots)
		}
	case 'R':
		if len(value) >= 4 {
			gsm, _ := strconv.Atoi(value[:2])
			satellites, _ := strconv.Atoi(value[2:4])
			decoded["gsmSignal"], decoded["satellites"] = gsm, satellites
			display = fmt.Sprintf("GSM信号：%d；GPS卫星：%d颗", gsm, satellites)
		}
	case 'W':
		if seconds, err := strconv.ParseUint(value, 16, 64); err == nil {
			decoded["seconds"] = seconds
			display = formatDuration(seconds)
		}
	case 'I':
		if stations, label, ok := decodeBaseStations(value, false); ok {
			decoded["stations"] = stations
			display = label
		}
	case 'Y':
		if stations, label, ok := decodeBaseStations(value, true); ok {
			decoded["stations"] = stations
			display = label
		}
	case 'Q':
		if hotspots, label, ok := decodeWiFi(value); ok {
			decoded["hotspots"] = hotspots
			display = label
		}
	case 'T':
		if battery, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			decoded["percent"] = battery
			display = fmt.Sprintf("%d%%", battery)
		}
	case 'N':
		networks := map[string]string{"00": "未知", "01": "GSM", "04": "LTE", "05": "CAT-M1（eMTC）", "06": "NB-IoT"}
		if label, exists := networks[value]; exists {
			decoded["network"] = label
			display = label
		}
	case 'V':
		display = value + "（设备电压原始值）"
	case 'K':
		display = decodeExtendedStatus(value)
	case 'J':
		if temperature, label, ok := decodeTemperature(value); ok {
			decoded["channel"], decoded["temperatureCelsius"] = temperature.channel, temperature.value
			decoded["negative"], decoded["sensorFault"], display = temperature.negative, temperature.sensorFault, label
		}
	case 'P':
		if labels, ok := decodeBlindArea(value); ok {
			decoded["conditions"], display = labels, strings.Join(labels, "｜")
		}
	case 'Z':
		if labels, ok := decodeDrivingStatus(value); ok {
			decoded["conditions"], display = labels, strings.Join(labels, "｜")
		}
	case 'X':
		parameters := decodeParameters(value)
		decoded["parameters"] = parameters
		if len(parameters) > 0 {
			display = strings.Join(parameters, "；")
		}
	}
	return name, display, decoded
}

func decodeGPS(value string) (map[string]interface{}, string, bool) {
	if len(value) < 34 {
		return nil, "", false
	}
	timePart, latPart, lngPart := value[:6], value[6:14], value[14:23]
	flags, err := strconv.ParseUint(value[23:24], 16, 8)
	if err != nil {
		return nil, "", false
	}
	lat, ok1 := degreesMinutes(latPart, 2)
	lng, ok2 := degreesMinutes(lngPart, 3)
	if !ok1 || !ok2 {
		return nil, "", false
	}
	if flags&0x02 == 0 {
		lat = -lat
	}
	if flags&0x04 == 0 {
		lng = -lng
	}
	speedCode, _ := strconv.Atoi(value[24:26])
	directionCode, _ := strconv.Atoi(value[26:28])
	datePart := value[28:34]
	timeText := datePart[4:6] + datePart[2:4] + datePart[:2] + timePart
	parsed, err := time.ParseInLocation("060102150405", timeText, beijingLocation)
	locationTime := timeText
	if err == nil {
		locationTime = parsed.Format("2006-01-02 15:04:05") + "（北京时间）"
	}
	positioned := flags&0x01 == 0
	result := map[string]interface{}{
		"latitude": lat, "longitude": lng, "positioned": positioned,
		"locationSource": map[bool]string{true: "GPS", false: "WiFi/基站"}[flags&0x08 == 0],
		"speedKmh":       float64(speedCode*2) * 1.852, "directionDegrees": directionCode * 10,
		"locationTime": locationTime, "flags": fmt.Sprintf("0x%X", flags),
	}
	label := fmt.Sprintf("时间：%s；纬度：%.6f；经度：%.6f；%s；速度：%.1f km/h；方向：%d°", locationTime, lat, lng, map[bool]string{true: "已定位", false: "未定位"}[positioned], float64(speedCode*2)*1.852, directionCode*10)
	return result, label, true
}

func degreesMinutes(value string, degreeDigits int) (float64, bool) {
	if len(value) <= degreeDigits {
		return 0, false
	}
	degree, err1 := strconv.ParseFloat(value[:degreeDigits], 64)
	minutesRaw, err2 := strconv.ParseFloat(value[degreeDigits:], 64)
	if err1 != nil || err2 != nil {
		return 0, false
	}
	minutes := minutesRaw / math.Pow10(len(value)-degreeDigits-2)
	return degree + minutes/60, true
}

func decodeBaseStations(value string, lte bool) ([]map[string]interface{}, string, bool) {
	if len(value) < 6 {
		return nil, "", false
	}
	count, err := strconv.Atoi(value[:1])
	if err != nil || count < 1 {
		return nil, "", false
	}
	cellLength := 4
	if lte {
		cellLength = 8
	}
	recordLength := 4 + cellLength + 2
	prefixLength := 6
	if lte && len(value) == 7+count*recordLength {
		prefixLength = 7
	}
	if len(value) != prefixLength+count*recordLength {
		return nil, "", false
	}
	mcc, mnc := value[1:4], value[4:prefixLength]
	stations := make([]map[string]interface{}, 0, count)
	for index := 0; index < count; index++ {
		start := prefixLength + index*recordLength
		signal, _ := strconv.Atoi(value[start+4+cellLength : start+recordLength])
		stations = append(stations, map[string]interface{}{"mcc": mcc, "mnc": mnc, "lac": value[start : start+4], "cellId": value[start+4 : start+4+cellLength], "signalDbm": signal - 113})
	}
	return stations, fmt.Sprintf("%d个%s基站；MCC=%s，MNC=%s", count, map[bool]string{true: "LTE", false: "GSM"}[lte], mcc, mnc), true
}

func decodeWiFi(value string) ([]map[string]interface{}, string, bool) {
	if len(value) < 2 {
		return nil, "", false
	}
	count, err := strconv.Atoi(value[:2])
	if err != nil || count < 0 || len(value) < 2+count*14 {
		return nil, "", false
	}
	hotspots := make([]map[string]interface{}, 0, count)
	labels := make([]string, 0, count)
	for index := 0; index < count; index++ {
		start := 2 + index*14
		macRaw := value[start : start+12]
		signalRaw := value[start+12 : start+14]
		signal, _ := strconv.Atoi(signalRaw)
		macParts := make([]string, 0, 6)
		for i := 0; i < 12; i += 2 {
			macParts = append(macParts, strings.ToLower(macRaw[i:i+2]))
		}
		mac := strings.Join(macParts, ":")
		dbm := signal - 113
		hotspots = append(hotspots, map[string]interface{}{"mac": mac, "signalDbm": dbm})
		labels = append(labels, fmt.Sprintf("#%d %s（%d dBm）", index+1, mac, dbm))
	}
	return hotspots, strings.Join(labels, "；"), true
}

func decodeExtendedStatus(value string) string {
	if len(value) < 5 {
		return value
	}
	labels := []string{}
	if nibble(value[0])&1 != 0 {
		labels = append(labels, "盲区补偿数据")
	}
	if nibble(value[0])&4 != 0 {
		labels = append(labels, "拆机报警")
	}
	if nibble(value[1])&1 != 0 {
		labels = append(labels, "追踪模式")
	}
	if nibble(value[3])&8 != 0 {
		labels = append(labels, "WiFi故障")
	}
	if len(labels) == 0 {
		return "无已知扩展状态"
	}
	return strings.Join(labels, "｜")
}

func decodeStatus(value string) ([]string, []string, string, bool) {
	if len(value) != 10 || !isHexText(value) {
		return nil, nil, "", false
	}
	status := make([]string, 0)
	alarms := make([]string, 0)
	if nibble(value[0])&0x04 != 0 {
		status = append(status, "GPS模块故障")
	}
	if nibble(value[1])&0x01 != 0 {
		status = append(status, "ACC开")
	} else {
		status = append(status, "ACC关")
	}
	if nibble(value[5])&0x04 != 0 {
		alarms = append(alarms, "震动报警")
	}
	labels := append(append([]string{}, status...), alarms...)
	if len(alarms) == 0 {
		labels = append(labels, "无已知报警")
	}
	return status, alarms, strings.Join(labels, "｜"), true
}

type temperatureData struct {
	channel     int
	value       float64
	negative    bool
	sensorFault bool
}

func decodeTemperature(value string) (temperatureData, string, bool) {
	if len(value) != 6 || !isHexText(value[:2]) {
		return temperatureData{}, "", false
	}
	channel, err := strconv.Atoi(value[:1])
	status, errStatus := strconv.ParseUint(value[1:2], 16, 8)
	number, errNumber := strconv.Atoi(value[2:])
	if err != nil || errStatus != nil || errNumber != nil {
		return temperatureData{}, "", false
	}
	negative := status&0x01 != 0
	temperature := float64(number) / 10
	if negative {
		temperature = -temperature
	}
	result := temperatureData{channel: channel, value: temperature, negative: negative, sensorFault: status&0x08 != 0}
	label := fmt.Sprintf("通道%d：%.1f℃", channel, temperature)
	if result.sensorFault {
		label += "（传感器故障）"
	}
	return result, label, true
}

func decodeBlindArea(value string) ([]string, bool) {
	if len(value) != 4 || !isHexText(value) {
		return nil, false
	}
	labels := make([]string, 0)
	definitions := []struct {
		index, bit int
		label      string
	}{
		{0, 0, "通讯受干扰"}, {0, 1, "无信号"},
		{2, 0, "平台未应答"}, {2, 1, "模块无响应"}, {2, 2, "SIM卡错误"}, {2, 3, "设备ID错误"},
		{3, 0, "网络注册失败"}, {3, 1, "PDP激活失败"}, {3, 2, "服务器IP连接失败"}, {3, 3, "域名连接失败"},
	}
	for _, definition := range definitions {
		if nibble(value[definition.index])&(1<<definition.bit) != 0 {
			labels = append(labels, definition.label)
		}
	}
	if len(labels) == 0 {
		labels = append(labels, "无盲区异常")
	}
	return labels, true
}

func decodeDrivingStatus(value string) ([]string, bool) {
	if len(value) < 3 || !isHexText(value[:3]) {
		return nil, false
	}
	labels := make([]string, 0)
	if nibble(value[1])&0x04 != 0 {
		labels = append(labels, "碰撞报警")
	}
	if nibble(value[1])&0x08 != 0 {
		labels = append(labels, "运动")
	} else {
		labels = append(labels, "静止")
	}
	return labels, true
}

func decodeParameters(value string) []string {
	parameters := make([]string, 0)
	for position := 0; position < len(value); {
		start := strings.IndexByte(value[position:], '(')
		if start < 0 {
			break
		}
		start += position
		end := strings.IndexByte(value[start+1:], ')')
		if end < 0 {
			break
		}
		end += start + 1
		parameters = append(parameters, value[start+1:end])
		position = end + 1
	}
	return parameters
}

func isHexText(value string) bool {
	if value == "" {
		return false
	}
	for index := range value {
		if !strings.ContainsRune("0123456789abcdefABCDEF", rune(value[index])) {
			return false
		}
	}
	return true
}

func nibble(value byte) int { parsed, _ := strconv.ParseInt(string(value), 16, 8); return int(parsed) }
func formatDuration(seconds uint64) string {
	return fmt.Sprintf("%d秒（%02d:%02d:%02d）", seconds, seconds/3600, (seconds%3600)/60, seconds%60)
}

func additionalName(code byte) string {
	names := map[byte]string{'A': "GPS定位数据", 'B': "车辆状态/报警", 'F': "速度数据", 'R': "信号与卫星", 'K': "扩展车辆状态", 'W': "系统运行时间", 'I': "GSM基站", 'Y': "LTE基站", 'T': "备用电池电量", 'J': "温度数据", 'X': "设备参数", 'P': "盲区数据类型", 'Z': "驾驶行为状态", 'N': "网络类型", 'Q': "WiFi热点", 'V': "设备电压"}
	if name, ok := names[code]; ok {
		return name
	}
	return fmt.Sprintf("附加信息（%c）", code)
}
func additionalDescription(raw []byte) string {
	if len(raw) < 2 {
		return "VDF附加信息"
	}
	return fmt.Sprintf("&%c类型附加信息，长度%d字节", raw[1], len(raw))
}
