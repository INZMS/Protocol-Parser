package jt808

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var beijingLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

func u16(data []byte) uint16 { return binary.BigEndian.Uint16(data) }
func u32(data []byte) uint32 { return binary.BigEndian.Uint32(data) }

func requireLength(data []byte, length int, message string) error {
	if len(data) < length {
		return fmt.Errorf("%s长度不足: 至少需要%d字节，实际%d字节", message, length, len(data))
	}
	return nil
}

func decodeBCDLoose(data []byte) string {
	var builder strings.Builder
	for _, value := range data {
		for _, digit := range []byte{value >> 4, value & 0x0f} {
			if digit <= 9 {
				builder.WriteByte('0' + digit)
			} else {
				builder.WriteByte('?')
			}
		}
	}
	return builder.String()
}

func normalizePhone(value string) string {
	// JT-808消息头固定使用BCD[6]保存12位号码；大陆11位手机号按协议在前补0。
	// 这里只移除一个协议填充位，不使用TrimLeft，避免破坏号码本身。
	if len(value) == 12 && strings.HasPrefix(value, "0") {
		return value[1:]
	}
	return value
}

func decodeTime(data []byte) (string, error) {
	if len(data) != 6 {
		return "", fmt.Errorf("JT-808时间字段必须为6字节BCD")
	}
	digits := decodeBCDLoose(data)
	if strings.Contains(digits, "?") {
		return "", fmt.Errorf("JT-808时间字段包含无效BCD数字: %s", hex.EncodeToString(data))
	}
	parsed, err := time.ParseInLocation("060102150405", digits, beijingLocation)
	if err != nil {
		return "", fmt.Errorf("JT-808时间字段无效: %w", err)
	}
	return parsed.Format("2006-01-02 15:04:05") + "（北京时间）", nil
}

func decodeGBK(data []byte) string {
	data = trimZero(data)
	decoded, _, err := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), data)
	if err != nil {
		return string(data)
	}
	return string(decoded)
}

func decodeUTF16BE(data []byte) string {
	data = trimZero(data)
	if len(data)%2 != 0 {
		return string(data)
	}
	words := make([]uint16, 0, len(data)/2)
	for index := 0; index < len(data); index += 2 {
		words = append(words, u16(data[index:index+2]))
	}
	return string(utf16.Decode(words))
}

func trimZero(data []byte) []byte {
	return []byte(strings.TrimRight(string(data), "\x00"))
}

func messageName(messageID uint16) string {
	names := map[uint16]string{
		0x0001: "终端通用应答", 0x8001: "平台通用应答", 0x0002: "终端心跳",
		0x0100: "终端注册", 0x8100: "终端注册应答", 0x0003: "终端注销",
		0x0102: "终端鉴权", 0x8103: "设置终端参数", 0x8105: "终端控制",
		0x0200: "位置信息汇报", 0x8201: "位置信息查询", 0x0201: "位置信息查询应答",
		0x8202: "临时位置跟踪控制", 0x0704: "盲区定位数据批量上传",
		0x8300: "文本信息下发", 0x6006: "上报文本消息",
	}
	if name, ok := names[messageID]; ok {
		return name
	}
	return fmt.Sprintf("未知消息0x%04X", messageID)
}
