# Aranet4 Go Reader

A lightweight, fast command-line tool written in Go to read data from Aranet4 CO2 sensors via Bluetooth Low Energy (BLE).

## Features

- 🚀 **Fast & Lightweight** - Single static binary with no runtime dependencies
- 📊 **Multiple Output Formats** - Human-readable text or JSON
- 🔌 **Direct BLE Connection** - No cloud services or internet required
- 🔓 **No Authentication Needed** - Works with Aranet4 devices that have Smart Home Integration enabled
- 🐧 **Cross-Platform** - Works on Linux and macOS
- 📦 **Zero Configuration** - Just provide the MAC address

## Requirements

- Aranet4 CO2 sensor (Home or Pro model)
- Firmware version v1.2.0 or greater
- Smart Home Integration enabled (check in the Aranet4 mobile app)
- Bluetooth adapter on your computer
- Go 1.16 or later (for building)

## Installation

### From Source

```bash
# Clone or download the project
mkdir aranet4-go
cd aranet4-go

# Initialize Go module
go mod init aranet4-go

# Save main.go (the code from this project)

# Install dependencies
go mod download

# Build
go build -o aranet4-go
```

### Pre-built Binaries

```bash
# Linux x64
GOOS=linux GOARCH=amd64 go build -o aranet4-go-linux-amd64

# Linux ARM64 (Raspberry Pi 4/5)
GOOS=linux GOARCH=arm64 go build -o aranet4-go-linux-arm64

# macOS (Intel) — requires CGO; must build on macOS
GOOS=darwin GOARCH=amd64 go build -o aranet4-go-darwin-amd64

# macOS (Apple Silicon) — requires CGO; must build on macOS
GOOS=darwin GOARCH=arm64 go build -o aranet4-go-darwin-arm64
```

## Finding Your Aranet4 MAC Address

You can find your device's MAC address using:

1. **Aranet4 Mobile App** - Check device settings
2. **Bluetooth Settings** - Pair with your phone and check paired devices
3. **Command Line Scan** (Linux):
   ```bash
   sudo hcitool lescan
   # or
   bluetoothctl scan on
   ```

The MAC address format is: `XX:XX:XX:XX:XX:XX` (e.g., `FC:5C:65:B7:84:94`)

## Usage

### Basic Reading

```bash
./aranet4-go -mac FC:5C:65:B7:84:94
```

**Sample Output:**
```
Scanning for Aranet4 device FC:5C:65:B7:84:94...
This may take up to 30 seconds. Make sure the device is awake.
Connecting...
Scanned 10 devices...
Found target device: FC:5C:65:B7:84:94 (Name: )
Found target device: FC:5C:65:B7:84:94 (Name: Aranet4 0A26A)
Connected to fc:5c:65:b7:84:94
Discovering services...
Reading data...
Connected: Aranet4 FC:5C:65:B7:84:94
CO2:         940 ppm
Temperature: 22.9 °C
Humidity:    30 %
Pressure:    1013.7 hPa
Battery:     94 %
Status:      GREEN
Interval:    300 s
Age:         4m 23s
```

**Note:** Values are color-coded by default based on health thresholds (green/yellow/red). Use `-m` flag to disable colors.

### JSON Output

```bash
./aranet4-go -mac FC:5C:65:B7:84:94 -json
```

**Sample Output:**
```json
{
  "co2": 940,
  "temperature": 22.9,
  "pressure": 1013.7,
  "humidity": 30,
  "battery": 94,
  "status": "GREEN",
  "status_code": 1,
  "interval": 300,
  "ago": 263
}
```

### Command-Line Options

```bash
Usage: aranet4-go -mac FC:5C:65:B7:84:94 [-json] [-debug] [-m]

Options:
  -mac string
        Aranet4 MAC address (required)
  -json
        Output in JSON format
  -debug
        Enable debug output (shows available BLE services and characteristics)
  -m    
        Disable colors (monochrome output)
```

### Color Coding

By default, sensor values are color-coded based on health thresholds:

**CO2 Levels (ppm):**
- 🟢 **GREEN**: < 1000 ppm (Good air quality)
- 🟡 **YELLOW**: 1000-1400 ppm (Average air quality, consider ventilation)
- 🔴 **RED**: > 1400 ppm (Poor air quality, ventilation needed)

**Temperature:**
- 🟢 **GREEN**: 18-24°C (64-75°F) (Comfortable)
- 🟡 **YELLOW**: 15-18°C or 24-27°C (59-64°F or 75-81°F) (Acceptable)
- 🔴 **RED**: < 15°C or > 27°C (< 59°F or > 81°F) (Uncomfortable)

**Humidity (%):**
- 🟢 **GREEN**: 30-45% (Optimal)
- 🟡 **YELLOW**: 46-70% (Acceptable)
- 🔴 **RED**: < 30% or > 70% (Too dry or too humid)

**Pressure (hPa):**
- 🟢 **GREEN**: 993-1033 hPa (Normal)
- 🟡 **YELLOW**: 980-992 hPa or 1034-1050 hPa (Low or high)
- 🔴 **RED**: < 980 or > 1050 hPa (Very low or very high)

**Battery (%):**
- 🟢 **GREEN**: > 30% (Good)
- 🟡 **YELLOW**: 15-30% (Low, consider charging)
- 🔴 **RED**: < 15% (Critical, needs charging)

Use the `-m` flag to disable colors for monochrome output or when piping to files.

## Understanding the Output

| Field | Description | Unit |
|-------|-------------|------|
| CO2 | Carbon dioxide concentration | ppm |
| Temperature | Ambient temperature | °C |
| Humidity | Relative humidity | % |
| Pressure | Atmospheric pressure | hPa |
| Battery | Battery level | % |
| Status | Air quality indicator | GREEN/YELLOW/RED |
| Interval | Measurement interval (configured in app) | seconds |
| Age | Time since last measurement | seconds |

### CO2 Status Levels

- **GREEN** (< 1000 ppm) - Good air quality
- **YELLOW** (1000-1400 ppm) - Average air quality, consider ventilation
- **RED** (> 1400 ppm) - Poor air quality, ventilation needed

## Linux Permissions

Bluetooth access requires special permissions:

### Option 1: Run with sudo (quick test)
```bash
sudo ./aranet4-go -mac FC:5C:65:B7:84:94
```

### Option 2: Grant capabilities (recommended)
```bash
sudo setcap 'cap_net_raw,cap_net_admin=eip' ./aranet4-go
./aranet4-go -mac FC:5C:65:B7:84:94
```

### Option 3: Add user to bluetooth group
```bash
sudo usermod -a -G bluetooth $USER
# Log out and log back in for changes to take effect
```

## Integration Examples

### Cron Job - Log Every 5 Minutes

```bash
# Add to crontab: crontab -e
*/5 * * * * /usr/local/bin/aranet4-go -mac FC:5C:65:B7:84:94 -json >> /var/log/aranet4.log
```

### Shell Script - CO2 Alert

```bash
#!/bin/bash
CO2=$(./aranet4-go -mac FC:5C:65:B7:84:94 -json | jq -r '.co2')

if [ $CO2 -gt 1000 ]; then
  echo "⚠️  High CO2 detected: $CO2 ppm"
  # Send notification, trigger automation, etc.
fi
```

### Python Integration

```python
import subprocess
import json

result = subprocess.run(
    ['./aranet4-go', '-mac', 'FC:5C:65:B7:84:94', '-json'],
    capture_output=True,
    text=True
)

data = json.loads(result.stdout)
print(f"CO2: {data['co2']} ppm")
```

### Home Assistant Integration

Use the ESPHome approach instead for Home Assistant (see below), but for command-line sensors:

```yaml
# configuration.yaml
sensor:
  - platform: command_line
    name: Aranet4 CO2
    command: '/usr/local/bin/aranet4-go -mac FC:5C:65:B7:84:94 -json'
    value_template: '{{ value_json.co2 }}'
    unit_of_measurement: 'ppm'
    scan_interval: 300
```

## Home Assistant Integration (ESPHome Method)

For a more robust Home Assistant integration, use an ESP32 device (like M5Stack Atom Lite) as a Bluetooth proxy:

1. Flash the ESP32 with ESPHome
2. Use the `esphome-aranet4` package
3. Place ESP32 within Bluetooth range of your Aranet4

See [stefanthoss/esphome-aranet4](https://github.com/stefanthoss/esphome-aranet4) for details.

## Troubleshooting

### "Device not found"
- Ensure Aranet4 is nearby (< 30 feet / 10 meters)
- Check battery level in Aranet4
- Verify MAC address is correct
- Enable Bluetooth on your computer

### "Aranet4 service not found"
- Update Aranet4 firmware to v1.2.0 or newer
- Enable "Smart Home Integration" in the Aranet4 mobile app
- Use `-debug` flag to see available services: `./aranet4-go -mac FC:5C:65:B7:84:94 -debug`

### "Permission denied"
- See [Linux Permissions](#linux-permissions) section above
- Ensure Bluetooth service is running: `sudo systemctl status bluetooth`

### Slow connection
- First connection takes 5-10 seconds (BLE scanning)
- Subsequent reads are faster if done within ~30 seconds

## Technical Details

### Protocol
- Uses Bluetooth Low Energy (BLE) GATT protocol
- Static UUIDs defined by SAF Tehnika (Aranet manufacturer)
- Service UUID: `fce0` (firmware v1.2.0+)
- Current readings characteristic: `f0cd3001-95da-4f4b-9ac8-aa55d312af0c`

### Data Format
The Aranet4 returns 13 bytes of data:
- Bytes 0-1: CO2 (uint16, little-endian) in ppm
- Bytes 2-3: Temperature (uint16 / 20.0) in °C
- Bytes 4-5: Pressure (uint16 / 10.0) in hPa
- Byte 6: Humidity (uint8) in %
- Byte 7: Battery (uint8) in %
- Byte 8: Status (1=green, 2=yellow, 3=red)
- Bytes 9-10: Interval (uint16) in seconds
- Bytes 11-12: Ago (uint16) in seconds

## Credits

Protocol information derived from:
- [Aranet4-Python](https://github.com/Anrijs/Aranet4-Python) by Anrijs Jargils
- [esphome-aranet4](https://github.com/stefanthoss/esphome-aranet4) by Stefan Thoss

## License

This project is provided as-is for personal and educational use.

## Contributing

Contributions welcome! Feel free to:
- Report issues
- Submit pull requests
- Share integration examples
- Improve documentation

## Related Projects

- [Aranet4-Python](https://github.com/Anrijs/Aranet4-Python) - Official Python library
- [esphome-aranet4](https://github.com/stefanthoss/esphome-aranet4) - ESPHome integration for Home Assistant
- [Aranet4 Home Assistant Integration](https://www.home-assistant.io/integrations/aranet/) - Official HA integration

---

**Made with ❤️ for better indoor air quality monitoring**