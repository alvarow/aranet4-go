package main

import (
	"math"
	"testing"
)

func TestParseAranet4Data(t *testing.T) {
	// CO2=940, Temp=22.9°C (458/20), Pressure=1013.7hPa (10137/10),
	// Humidity=30, Battery=94, StatusCode=1, Interval=300, Ago=263
	buf := []byte{
		0xAC, 0x03, // CO2 = 940
		0xCA, 0x01, // Temp raw = 458 → 22.9°C
		0x99, 0x27, // Pressure raw = 10137 → 1013.7 hPa
		30,         // Humidity
		94,         // Battery
		1,          // StatusCode
		0x2C, 0x01, // Interval = 300
		0x07, 0x01, // Ago = 263
	}

	d := parseAranet4Data(buf)

	if d.CO2 != 940 {
		t.Errorf("CO2: got %d, want 940", d.CO2)
	}
	if math.Abs(float64(d.Temperature)-22.9) > 0.01 {
		t.Errorf("Temperature: got %f, want 22.9", d.Temperature)
	}
	if math.Abs(float64(d.Pressure)-1013.7) > 0.1 {
		t.Errorf("Pressure: got %f, want 1013.7", d.Pressure)
	}
	if d.Humidity != 30 {
		t.Errorf("Humidity: got %d, want 30", d.Humidity)
	}
	if d.Battery != 94 {
		t.Errorf("Battery: got %d, want 94", d.Battery)
	}
	if d.StatusCode != 1 {
		t.Errorf("StatusCode: got %d, want 1", d.StatusCode)
	}
	if d.Interval != 300 {
		t.Errorf("Interval: got %d, want 300", d.Interval)
	}
	if d.Ago != 263 {
		t.Errorf("Ago: got %d, want 263", d.Ago)
	}
	if d.Status != "GREEN" {
		t.Errorf("Status: got %s, want GREEN", d.Status)
	}
}

func TestParseAranet4DataHighCO2(t *testing.T) {
	// CO2=1500 should produce RED status
	buf := []byte{
		0xDC, 0x05, // CO2 = 1500
		0x00, 0x00, // Temp = 0
		0x00, 0x00, // Pressure = 0
		45,         // Humidity (GREEN)
		50,         // Battery
		3,          // StatusCode
		0x00, 0x00, // Interval
		0x00, 0x00, // Ago
	}
	d := parseAranet4Data(buf)
	if d.Status != "RED" {
		t.Errorf("Status: got %s, want RED", d.Status)
	}
}

func TestCO2AirStatus(t *testing.T) {
	tests := []struct {
		co2  uint16
		want string
	}{
		{0, "GREEN"},
		{999, "GREEN"},
		{1000, "YELLOW"},
		{1400, "YELLOW"},
		{1401, "RED"},
		{2000, "RED"},
	}
	for _, tt := range tests {
		if got := co2AirStatus(tt.co2); got != tt.want {
			t.Errorf("co2AirStatus(%d) = %s, want %s", tt.co2, got, tt.want)
		}
	}
}

func TestHumidityAirStatus(t *testing.T) {
	tests := []struct {
		h    uint8
		want string
	}{
		{29, "RED"},
		{30, "GREEN"},
		{45, "GREEN"},
		{46, "YELLOW"},
		{70, "YELLOW"},
		{71, "RED"},
	}
	for _, tt := range tests {
		if got := humidityAirStatus(tt.h); got != tt.want {
			t.Errorf("humidityAirStatus(%d) = %s, want %s", tt.h, got, tt.want)
		}
	}
}

func TestTempAirStatus(t *testing.T) {
	tests := []struct {
		temp float32
		want string
	}{
		{14.9, "RED"},
		{15.0, "YELLOW"},
		{17.9, "YELLOW"},
		{18.0, "GREEN"},
		{24.0, "GREEN"},
		{24.1, "YELLOW"},
		{27.0, "YELLOW"},
		{27.1, "RED"},
	}
	for _, tt := range tests {
		if got := tempAirStatus(tt.temp); got != tt.want {
			t.Errorf("tempAirStatus(%.1f) = %s, want %s", tt.temp, got, tt.want)
		}
	}
}

func TestPressureAirStatus(t *testing.T) {
	tests := []struct {
		pressure float32
		want     string
	}{
		{979.9, "RED"},
		{980.0, "YELLOW"},
		{992.9, "YELLOW"},
		{993.0, "GREEN"},
		{1033.0, "GREEN"},
		{1033.1, "YELLOW"},
		{1050.0, "YELLOW"},
		{1050.1, "RED"},
	}
	for _, tt := range tests {
		if got := pressureAirStatus(tt.pressure); got != tt.want {
			t.Errorf("pressureAirStatus(%.1f) = %s, want %s", tt.pressure, got, tt.want)
		}
	}
}

func TestBatteryAirStatus(t *testing.T) {
	tests := []struct {
		b    uint8
		want string
	}{
		{0, "RED"},
		{14, "RED"},
		{15, "YELLOW"},
		{30, "YELLOW"},
		{31, "GREEN"},
		{100, "GREEN"},
	}
	for _, tt := range tests {
		if got := batteryAirStatus(tt.b); got != tt.want {
			t.Errorf("batteryAirStatus(%d) = %s, want %s", tt.b, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		secs uint16
		want string
	}{
		{0, "0s"},
		{59, "59s"},
		{60, "1m"},
		{90, "1m 30s"},
		{263, "4m 23s"},
		{300, "5m"},
		{3600, "60m"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.secs); got != tt.want {
			t.Errorf("formatDuration(%d) = %s, want %s", tt.secs, got, tt.want)
		}
	}
}

func TestIsValidMACAddress(t *testing.T) {
	valid := []string{
		"FC:5C:65:B7:84:94",
		"00:00:00:00:00:00",
		"FF:FF:FF:FF:FF:FF",
	}
	for _, mac := range valid {
		if !isValidMACAddress(mac) {
			t.Errorf("isValidMACAddress(%q) = false, want true", mac)
		}
	}

	invalid := []string{
		"fc:5c:65:b7:84:94",    // lowercase
		"FC-5C-65-B7-84-94",    // dashes (pre-normalization)
		"FC:5C:65:B7:84",       // too short
		"FC:5C:65:B7:84:94:00", // too long
		"",
		"not-a-mac",
	}
	for _, mac := range invalid {
		if isValidMACAddress(mac) {
			t.Errorf("isValidMACAddress(%q) = true, want false", mac)
		}
	}
}
