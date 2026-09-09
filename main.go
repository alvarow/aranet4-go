package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

const (
	aranetServiceUUID   = "fce0"
	currentReadingsUUID = "f0cd3001-95da-4f4b-9ac8-aa55d312af0c"
	scanTimeout         = 45 * time.Second
)

var (
	Version    = "dev"
	BuildTime  = "unknown"
	GitCommit  = "unknown"
	DefaultMAC = ""

	macRegex = regexp.MustCompile(`^([0-9A-F]{2}:){5}[0-9A-F]{2}$`)
)

func isValidMACAddress(mac string) bool {
	return macRegex.MatchString(mac)
}

// Aranet4Data represents a reading from the Aranet4 sensor
type Aranet4Data struct {
	CO2         uint16  `json:"co2"`         // ppm
	Temperature float32 `json:"temperature"` // °C
	Pressure    float32 `json:"pressure"`    // hPa
	Humidity    uint8   `json:"humidity"`    // %
	Battery     uint8   `json:"battery"`     // %
	Status      string  `json:"status"`      // GREEN, YELLOW, RED
	StatusCode  uint8   `json:"status_code"` // 1=green, 2=yellow, 3=red
	Interval    uint16  `json:"interval"`    // seconds
	Ago         uint16  `json:"ago"`         // seconds since last measurement
}

func main() {
	macAddr := flag.String("mac", DefaultMAC, "Aranet4 MAC address")
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	debug := flag.Bool("debug", false, "Enable debug output")
	monochrome := flag.Bool("m", false, "Disable colors (monochrome output)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.BoolVar(versionFlag, "v", false, "Print version and exit")
	help := flag.Bool("h", false, "Show help")
	flag.BoolVar(help, "help", false, "Show help")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("aranet4-go v%s (commit: %s, built: %s)\n", Version, GitCommit, BuildTime)
		os.Exit(0)
	}

	if *help || *macAddr == "" {
		fmt.Printf("Aranet4 Go Reader v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		if DefaultMAC != "" {
			fmt.Printf("Usage: aranet4-go [-mac %s] [-json] [-debug] [-m]\n", DefaultMAC)
		} else {
			fmt.Println("Usage: aranet4-go -mac <MAC_ADDRESS> [-json] [-debug] [-m]")
			fmt.Println("Note: Build with DEFAULT-MAC-ADDR file to set a default MAC address")
		}
		flag.PrintDefaults()
		if *macAddr == "" {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *debug {
		fmt.Fprintf(os.Stderr, "Aranet4 Go Reader v%s\n", Version)
		fmt.Fprintf(os.Stderr, "Build Time: %s\n", BuildTime)
		fmt.Fprintf(os.Stderr, "Git Commit: %s\n", GitCommit)
	}

	data, err := readAranet4(*macAddr, *debug)
	if err != nil {
		log.Fatalf("Error reading Aranet4: %v", err)
	}

	if *jsonOutput {
		outputJSON(data)
	} else {
		outputHuman(data, *macAddr, *debug, !*monochrome)
	}
}

func readAranet4(macAddr string, debug bool) (*Aranet4Data, error) {
	d, err := dev.NewDevice("default")
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}
	ble.SetDefaultDevice(d)

	macAddr = strings.ToUpper(strings.ReplaceAll(macAddr, "-", ":"))
	if !isValidMACAddress(macAddr) {
		return nil, fmt.Errorf("invalid MAC address format: %s", macAddr)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Scanning for Aranet4 device %s...\n", macAddr)
		fmt.Fprintf(os.Stderr, "This may take up to 30 seconds. Make sure the device is awake.\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	var foundCount int
	filter := func(adv ble.Advertisement) bool {
		foundCount++
		if debug && foundCount%10 == 0 {
			fmt.Fprintf(os.Stderr, "Scanned %d devices...\n", foundCount)
		}

		addr := strings.ToUpper(adv.Addr().String())
		if addr == macAddr {
			if debug {
				fmt.Fprintf(os.Stderr, "Found target device: %s (Name: %s)\n", addr, adv.LocalName())
			}
			return true
		}
		return false
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Connecting...\n")
	}
	client, err := ble.Connect(ctx, filter)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout: device not found after %v. Scanned %d devices. Is the Aranet4 nearby and awake?", scanTimeout, foundCount)
		}
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = client.CancelConnection() }()

	if debug {
		fmt.Fprintf(os.Stderr, "Connected to %s\n", client.Addr().String())
		fmt.Fprintf(os.Stderr, "Discovering services...\n")
	}

	serviceUUID, err := ble.Parse(aranetServiceUUID)
	if err != nil {
		return nil, fmt.Errorf("invalid service UUID: %w", err)
	}

	charUUID, err := ble.Parse(currentReadingsUUID)
	if err != nil {
		return nil, fmt.Errorf("invalid characteristic UUID: %w", err)
	}

	profile, err := client.DiscoverProfile(true)
	if err != nil {
		return nil, fmt.Errorf("failed to discover profile: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Available services:\n")
		for _, service := range profile.Services {
			fmt.Fprintf(os.Stderr, "  Service: %s\n", service.UUID.String())
			for _, char := range service.Characteristics {
				fmt.Fprintf(os.Stderr, "    Characteristic: %s (Properties: %v)\n", char.UUID.String(), char.Property)
			}
		}
	}

	var targetChar *ble.Characteristic
	for _, service := range profile.Services {
		if service.UUID.Equal(serviceUUID) {
			for _, char := range service.Characteristics {
				if char.UUID.Equal(charUUID) {
					targetChar = char
					break
				}
			}
		}
	}

	if targetChar == nil {
		return nil, fmt.Errorf("Aranet4 characteristic not found")
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Reading data...\n")
	}

	buf, err := client.ReadCharacteristic(targetChar)
	if err != nil {
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "insufficient") {
			fmt.Fprintf(os.Stderr, "Authentication required. Make sure 'Smart Home Integration' is enabled in the Aranet4 app.\n")
			fmt.Fprintf(os.Stderr, "If enabled, try pairing the device first with: bluetoothctl pair %s\n", macAddr)
		}
		return nil, fmt.Errorf("failed to read characteristic: %w", err)
	}

	if len(buf) != 13 {
		return nil, fmt.Errorf("invalid data length: got %d bytes, expected exactly 13", len(buf))
	}

	return parseAranet4Data(buf), nil
}

func co2AirStatus(co2 uint16) string {
	if co2 > 1400 {
		return "RED"
	}
	if co2 >= 1000 {
		return "YELLOW"
	}
	return "GREEN"
}

func humidityAirStatus(h uint8) string {
	if h < 30 || h > 70 {
		return "RED"
	}
	if h >= 46 {
		return "YELLOW"
	}
	return "GREEN"
}

func tempAirStatus(temp float32) string {
	if temp < 15 || temp > 27 {
		return "RED"
	}
	if temp < 18 || temp > 24 {
		return "YELLOW"
	}
	return "GREEN"
}

func pressureAirStatus(pressure float32) string {
	if pressure < 980 || pressure > 1050 {
		return "RED"
	}
	if pressure < 993 || pressure > 1033 {
		return "YELLOW"
	}
	return "GREEN"
}

func batteryAirStatus(b uint8) string {
	if b < 15 {
		return "RED"
	}
	if b <= 30 {
		return "YELLOW"
	}
	return "GREEN"
}

func colorStyle(status string, useColors bool) lipgloss.Style {
	if !useColors {
		return lipgloss.NewStyle()
	}
	switch status {
	case "GREEN":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	case "YELLOW":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	}
}

func formatDuration(secs uint16) string {
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	m := secs / 60
	s := secs % 60
	if s == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

// parseAranet4Data decodes the 13-byte response from Aranet4.
// Format: https://github.com/Anrijs/Aranet4-Python/blob/master/aranet4/client.py#L317-L333
func parseAranet4Data(data []byte) *Aranet4Data {
	result := &Aranet4Data{
		CO2:         uint16(data[0]) | uint16(data[1])<<8,
		Temperature: float32(uint16(data[2])|uint16(data[3])<<8) / 20.0,
		Pressure:    float32(uint16(data[4])|uint16(data[5])<<8) / 10.0,
		Humidity:    data[6],
		Battery:     data[7],
		StatusCode:  data[8],
		Interval:    uint16(data[9]) | uint16(data[10])<<8,
		Ago:         uint16(data[11]) | uint16(data[12])<<8,
	}

	co2S := co2AirStatus(result.CO2)
	humS := humidityAirStatus(result.Humidity)

	if co2S == "RED" || humS == "RED" {
		result.Status = "RED"
	} else if co2S == "YELLOW" || humS == "YELLOW" {
		result.Status = "YELLOW"
	} else {
		result.Status = "GREEN"
	}

	return result
}

func outputJSON(data *Aranet4Data) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("JSON encoding failed: %v", err)
	}
}

func outputHuman(data *Aranet4Data, macAddr string, debug bool, useColors bool) {
	if debug {
		fmt.Printf("Connected: Aranet4 %s\n", macAddr)
	}

	fmt.Printf("CO2:         %s\n", colorStyle(co2AirStatus(data.CO2), useColors).Render(fmt.Sprintf("%d ppm", data.CO2)))
	fmt.Printf("Temperature: %s\n", colorStyle(tempAirStatus(data.Temperature), useColors).Render(fmt.Sprintf("%.1f °C", data.Temperature)))
	fmt.Printf("Humidity:    %s\n", colorStyle(humidityAirStatus(data.Humidity), useColors).Render(fmt.Sprintf("%d %%", data.Humidity)))
	fmt.Printf("Pressure:    %s\n", colorStyle(pressureAirStatus(data.Pressure), useColors).Render(fmt.Sprintf("%.1f hPa", data.Pressure)))
	fmt.Printf("Battery:     %s\n", colorStyle(batteryAirStatus(data.Battery), useColors).Render(fmt.Sprintf("%d %%", data.Battery)))
	fmt.Printf("Status:      %s\n", colorStyle(data.Status, useColors).Render(data.Status))
	fmt.Printf("Interval:    %d s\n", data.Interval)
	fmt.Printf("Age:         %s\n", formatDuration(data.Ago))
}
