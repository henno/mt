# mt - Mikrotik CLI Tool

A command-line tool for executing Mikrotik RouterOS commands via API or SSH.

## Installation

```bash
go build -o mt .
```

For SSH mode, `sshpass` is required:
```bash
brew install hudochenkov/sshpass/sshpass  # macOS
apt install sshpass                        # Debian/Ubuntu
```

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
MT_HOST=192.168.88.1
MT_USER=admin
MT_PASSWORD=yourpassword
MT_PORT=8728
#MT_USE_TLS=true
#MT_USE_SSH=true
```

Enable API on your Mikrotik: `/ip service enable api`

## Usage

### API Mode (default)

Uses RouterOS API protocol (port 8728, or 8729 with TLS):

```bash
./mt -c '/system/resource/print'
./mt -c '/interface/print'
./mt -c '/ip/service/print ?name=api'
./mt -c '/ip/service/set =.id=*0 =address=10.0.0.0/24'
```

### SSH Mode

Uses SSH for CLI commands (port 22). Supports `export` and other CLI-only commands:

```bash
./mt -ssh -c '/user export'
./mt -ssh -c '/system resource print'
./mt -ssh -c '/interface print where type=ether'
```

### Filtering

- **API mode**: Use `?` prefix (e.g., `?name=api`)
- **SSH mode**: Use `where` keyword (e.g., `where type=ether`)

## macOS: Copying the binary

When copying the binary to another folder, macOS may block it due to security restrictions. Two solutions:

**Option 1: Symlink (recommended)**
```bash
ln -sf /path/to/mt /destination/mt
```

**Option 2: Re-sign after copying**
```bash
cp mt /destination/
codesign -s - /destination/mt
```
