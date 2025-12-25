# mt - Mikrotik CLI Tool

A simple command-line tool for executing Mikrotik RouterOS API commands.

## Installation

```bash
go build -o mt .
```

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
MT_HOST=192.168.88.1
MT_USER=admin
MT_PASSWORD=yourpassword
MT_PORT=8728
#MT_USE_TLS=true
```

Enable API on your Mikrotik: `/ip service enable api`

## Usage

```bash
./mt -c '/system/resource/print'
./mt -c '/interface/print'
./mt -c '/ip/service/print ?name=api'
./mt -c '/ip/service/set =.id=*0 =address=10.0.0.0/24'
```

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
