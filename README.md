# OSC Checker

A professional OSC (Open Sound Control) sender and receiver application built with Go and Fyne. This tool provides an easy way to test and debug OSC communication within local networks, featuring a clean Protokol-inspired user interface.

## Features

### 🎛️ OSC Sender
- **Target Configuration**: Set IP address and port for OSC destinations
- **Custom OSC Addresses**: Send messages to any OSC address path
- **Multiple Argument Types**: Support for int, float, string, and bool arguments
- **Dynamic Arguments**: Add/remove arguments as needed
- **Send History**: Track your sent messages

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

1. **Configure Target**:
   - Enter the target IP address (default: 127.0.0.1)
   - Set the target port (default: 7000)

2. **Set OSC Address**:
   - Enter the OSC address path (e.g., `/test/sample`)

3. **Add Arguments** (optional):
   - Click "Add Argument" to add parameters
   - Select argument type: int, float, string, or bool
   - Enter the value
   - Remove arguments with the "Remove" button

4. **Send Message**:
   - Click "Send" to transmit the OSC message
   - Check the send history below

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

The application uses a `config.yaml` file for default settings:

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

receiver:
  default_port: 7000
  window:
    width: 1000
    height: 700
    title: "OSC Receiver"
  max_log_entries: 1000
```

## Use Cases

### Development & Testing
- **API Testing**: Test OSC-based applications and plugins
- **Network Debugging**: Verify OSC message transmission across networks
- **Protocol Validation**: Ensure correct OSC message formatting

### Live Performance
- **Signal Monitoring**: Monitor OSC traffic in real-time during performances
- **Connection Verification**: Verify connections between different software/hardware

### Education
- **OSC Learning**: Understand OSC protocol behavior
- **Network Communication**: Learn about UDP-based communication protocols

## Technical Details

- **Framework**: Fyne v2 (Cross-platform GUI)
- **OSC Library**: github.com/hypebeast/go-osc
- **Configuration**: YAML-based configuration
- **Platform Support**: Windows, macOS, Linux
- **Language**: Go 1.19+

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

2. **Port already in use**:
   - Change the receiver port number
   - Check if another application is using the port

3. **Messages not filtered correctly**:
   - Ensure correct filter syntax (use `*` for wildcards)
   - Check for typos in the filter input

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
