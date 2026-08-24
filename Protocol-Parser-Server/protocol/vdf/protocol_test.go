package vdf

import (
	"encoding/hex"
	"math"
	"testing"
)

const locationExample = "*HQ201861803178999002,BA&A1111182233396811356521160016090718&B0100000000&F0000&R1807&W00000029&I54600024950E115124950E1254249513B3422495122335262310D735&K40100&T95#"

func TestParseLocationExample(t *testing.T) {
	result, err := New().Parse([]byte(locationExample))
	if err != nil {
		t.Fatal(err)
	}
	if result.Protocol != "VDF" || result.MessageID != "BA" || result.MessageName != "定时定位上报" {
		t.Fatalf("unexpected result: %#v", result)
	}
	data := result.Data.(map[string]interface{})
	items := data["additional"].([]map[string]interface{})
	gps := items[0]
	if math.Abs(gps["latitude"].(float64)-22.5566133333) > 0.000001 {
		t.Fatalf("latitude=%v", gps["latitude"])
	}
	if math.Abs(gps["longitude"].(float64)-113.9420183333) > 0.000001 {
		t.Fatalf("longitude=%v", gps["longitude"])
	}
}

func TestParseOnlineMessage(t *testing.T) {
	result, err := New().Parse([]byte("*HQ20013800138000,AY#"))
	if err != nil {
		t.Fatal(err)
	}
	if result.MessageID != "AY" || result.MessageName != "设备上线提示" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestRejectInvalidFrame(t *testing.T) {
	_, err := New().Parse([]byte("HQ20013800138000,AY"))
	if err == nil {
		t.Fatal("expected invalid frame error")
	}
}

func TestRejectInvalidReplyAttribute(t *testing.T) {
	_, err := New().Parse([]byte("*HQ20X13800138000,AY#"))
	if err == nil {
		t.Fatal("expected invalid reply attribute error")
	}
}

func TestDecodeThreeDigitLTEMNC(t *testing.T) {
	stations, _, ok := decodeBaseStations("1"+"460"+"011"+"1234"+"ABCDEF01"+"50", true)
	if !ok || len(stations) != 1 {
		t.Fatalf("failed to decode LTE station: %#v", stations)
	}
	if stations[0]["mnc"] != "011" || stations[0]["cellId"] != "ABCDEF01" || stations[0]["signalDbm"] != -63 {
		t.Fatalf("unexpected LTE station: %#v", stations[0])
	}
}

func TestDecodeStatusAndTemperature(t *testing.T) {
	status, alarms, _, ok := decodeStatus("4100040000")
	if !ok || len(status) != 2 || len(alarms) != 1 || alarms[0] != "震动报警" {
		t.Fatalf("unexpected status: %#v %#v", status, alarms)
	}
	temperature, _, ok := decodeTemperature("050258")
	if !ok || temperature.channel != 0 || math.Abs(temperature.value-(-25.8)) > 0.001 {
		t.Fatalf("unexpected temperature: %#v", temperature)
	}
}

func TestDecodeWiFi(t *testing.T) {
	hotspots, _, ok := decodeWiFi("01AABBCCDDEEFF62")
	if !ok || len(hotspots) != 1 || hotspots[0]["mac"] != "aa:bb:cc:dd:ee:ff" || hotspots[0]["signalDbm"] != -51 {
		t.Fatalf("unexpected hotspots: %#v", hotspots)
	}
}

func TestRawIsHex(t *testing.T) {
	result, err := New().Parse([]byte("*HQ20013800138000,AY#"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hex.DecodeString(result.Raw); err != nil {
		t.Fatalf("raw is not hex: %v", err)
	}
}
