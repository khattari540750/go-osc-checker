# OSC Checker

A professional OSC (Open Sound Control) sender and receiver application built with Go and Fyne. This tool provides an easy way to test and debug OSC communication within local networks, featuring a clean modern user interface with support for multiple sender configurations.

## Features

### 🎛️ OSC Sender
- **Multiple Targets**: Configure multiple OSC destinations from YAML config
- **Target Configuration**: Set IP address and port for each OSC destination
- **Custom OSC Addresses**: Send messages to any OSC address path
- **Preset Arguments**: Pre-configured argument templates with descriptions
- **Multiple Argument Types**: Support for int, float, string, and bool arguments
- **Dynamic Arguments**: Add/remove arguments as needed with intuitive ＋/✕ buttons
- **Send History**: Track your sent messages with timestamps
- **Clean UI**: Large, accessible buttons and streamlined interface

### 📡 OSC Receiver
- **Real-time Monitoring**: Live display of incoming OSC messages
- **Advanced Filtering**: 
  - Wildcard support (`/test*` matches addresses starting with `/test`)
  - Partial matching (`/tet` matches addresses containing `/test`)
  - Real-time filter updates
- **Session Management**: 
  - Auto-clear on start for clean test sessions
  - Manual clear functionality
- **Professional UI**: 
  - Protokol-style design
  - Status indicators with visual feedback
  - Optimized button placement and sizing
- **Message Log**: Timestamped message history with filtering
- **Export Functionality**: Save logs to text files

## Installation

### Prerequisites
- Go 1.19 or later
- Git

### Build from Source
```bash
git clone <repository-url>
cd osc-checker
go mod tidy
go build -o osc-checker main.go
```

### Run
```bash
./osc-checker
```

## Usage

### Starting the Application
When you run the application, two windows will open:
1. **OSC Sender** - For sending OSC messages
2. **OSC Receiver** - For monitoring incoming OSC messages

### OSC Sender Usage

The sender interface shows multiple configured targets from your config.yaml file.

1. **Select Target**:
   - Each target shows its name (e.g., "TestServer", "LiveServer")
   - IP, Port, and OSC Address fields are pre-configured but editable
   - Large "Send" button is positioned next to the target name for easy access

2. **Configure Message**:
   - **IP**: Target IP address (default from config)
   - **Port**: Target port number (default from config)  
   - **OSC Addr**: OSC address path (default from config)

3. **Manage Arguments**:
   - Pre-configured arguments are loaded from config with defaults
   - **Add**: Click the ＋ button to add new arguments
   - **Remove**: Click the ✕ button to remove arguments
   - **Types**: Select from int, float, string, or bool
   - **Values**: Enter values directly in the fields

4. **Send Message**:
   - Click the prominent "Send" button next to the target name
   - Messages are sent immediately
   - Check the send history at the bottom of the window

### OSC Receiver Usage

1. **Configure Receiver**:
   - Set the listening port (default: 7000)
   - Port field accepts up to 5-digit port numbers

2. **Start Receiving**:
   - Click the prominent "Start" button
   - Status will change to "Receiving..." with a green indicator
   - The message log will automatically clear for a fresh session

3. **Filter Messages**:
   - Use the "Address Filter" field for real-time filtering
   - Examples:
     - `/test*` - Shows messages starting with `/test`
     - `/tet` - Shows messages containing `/tet`
     - Empty - Shows all messages

4. **Manage Logs**:
   - **Clear**: Manual clear button next to "Message Log" header
   - **Save**: Export current log to a timestamped text file
   - Real-time message counter shows total received messages

5. **Stop Receiving**:
   - Click "Stop" to halt message reception
   - Status will change to "Stopped" with a red indicator

## Configuration

The application uses a `config.yaml` file for multiple sender targets and default settings:

```yaml
app:
  name: "OSC Checker"
  version: "1.0.0"

sender:
  default_host: "127.0.0.1"
  default_port: 7000
  default_address: "/test"
  window:
    width: 900
    height: 600
    title: "OSC Sender"
  list:
    - name: "TestServer"
      host: "127.0.0.1"
      port: 7000
      address: "/test"
      arguments:
        - type: "int"
          default_value: "1"
          description: "Test value"
        - type: "float"
          default_value: "1.5"
          description: "Volume level"
    - name: "LiveServer"
      host: "192.168.1.100"
      port: 8000
      address: "/live/trigger"
      arguments:
        - type: "string"
          default_value: "trigger"
          description: "Command"

receiver:
  default_port: 7000
  window:
    width: 1000
    height: 700
    title: "OSC Receiver"
  max_log_entries: 1000
```

### Configuration Features
- **Multiple Targets**: Define multiple sender configurations
- **Preset Arguments**: Pre-configure arguments with default values and descriptions
- **Flexible Setup**: Each target can have different hosts, ports, and addresses
- **UI Customization**: Window sizes and titles are configurable

## Use Cases

### Development & Testing
- **Multi-Target Testing**: Test multiple OSC destinations simultaneously
- **API Testing**: Test OSC-based applications and plugins with preset configurations
- **Network Debugging**: Verify OSC message transmission across networks
- **Protocol Validation**: Ensure correct OSC message formatting

### Live Performance
- **Signal Monitoring**: Monitor OSC traffic in real-time during performances
- **Multi-Device Control**: Send OSC messages to multiple devices/applications
- **Connection Verification**: Verify connections between different software/hardware

### Education & Research
- **OSC Learning**: Understand OSC protocol behavior with multiple examples
- **Network Communication**: Learn about UDP-based communication protocols
- **Configuration Management**: Learn YAML-based application configuration

## Technical Details

- **Framework**: Fyne v2 (Cross-platform GUI)
- **OSC Library**: github.com/hypebeast/go-osc
- **Configuration**: YAML-based configuration with multiple sender support
- **Platform Support**: Windows, macOS, Linux
- **Language**: Go 1.19+
- **UI Features**: Large buttons, intuitive symbols (＋/✕), streamlined layout

## Message Format

The receiver displays messages in the following format:
```
TIME     │ ADDRESS              │ VALUES
15:04:05 │ /test/sample         │ 1.0, hello, true
```

## Filter Examples

| Filter Input | Matches | Description |
|-------------|---------|-------------|
| `/test*` | `/test/anything` | Wildcard: starts with `/test` |
| `/tet` | `/test/sample` | Partial: contains `/tet` |
| `/osc/volume` | `/osc/volume` | Exact: matches exactly |
| (empty) | All messages | Show all received messages |

## Troubleshooting

### Common Issues

1. **No messages received**:
   - Check if the sender and receiver are using the same port
   - Verify firewall settings
   - Ensure the target IP address is correct
   - Verify config.yaml has valid sender configurations

2. **Port already in use**:
   - Change the receiver port number
   - Check if another application is using the port
   - Try different ports for different sender targets

3. **Messages not filtered correctly**:
   - Ensure correct filter syntax (use `*` for wildcards)
   - Check for typos in the filter input

4. **Configuration issues**:
   - Verify config.yaml syntax is correct
   - Check that all required fields are present in sender list
   - Ensure argument types are valid (int, float, string, bool)

5. **UI not responding**:
   - Check console output for error messages
   - Verify all dependencies are installed
   - Try rebuilding the application

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by the Protokol OSC monitoring interface
- Built with the excellent Fyne GUI framework
- Uses the reliable go-osc library for OSC communication
