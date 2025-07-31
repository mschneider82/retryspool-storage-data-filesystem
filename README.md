# RetrySpool Filesystem Data Storage

Filesystem implementation for RetrySpool data storage backend.

## Overview

This package provides a filesystem-based implementation of the RetrySpool data storage interface. It stores message data as individual files on the filesystem with directory sharding for performance.

## Installation

```bash
go get schneider.vip/retryspool/storage/data/filesystem
```

## Usage

```go
import (
    datastorage "schneider.vip/retryspool/storage/data"
    "schneider.vip/retryspool/storage/data/filesystem"
)

// Create factory
factory := filesystem.NewFactory("/path/to/data/storage")

// Register with registry (optional)
// Factory is ready to use directly

// Create backend
backend, err := factory.Create()
if err != nil {
    panic(err)
}
defer backend.Close()

// Store data
size, err := backend.StoreData(ctx, "msg-123", dataReader)
if err != nil {
    panic(err)
}

// Read data
reader, err := backend.GetDataReader(ctx, "msg-123")
if err != nil {
    panic(err)
}
defer reader.Close()
```

## Directory Structure

The filesystem backend organizes data files using directory sharding:

```
/path/to/data/storage/
├── ab/
│   ├── abcd1234.data
│   └── ab9876543.data
├── cd/
│   ├── cd123456.data
│   └── cdef7890.data
└── misc/
    └── x.data          # Messages with IDs shorter than 2 chars
```

## Features

- **Directory Sharding**: Uses first 2 characters of message ID for directory organization
- **Efficient I/O**: Direct file system operations for optimal performance
- **Automatic Cleanup**: Removes empty directories when deleting files
- **Thread Safety**: Safe for concurrent access
- **Large File Support**: Efficient handling of large message data

## Performance Characteristics

- **Read Performance**: Direct file access, very fast
- **Write Performance**: Direct file writes, very fast
- **Storage Efficiency**: No overhead, stores only the actual data
- **Scalability**: Directory sharding prevents filesystem bottlenecks

## Configuration

The only configuration needed is the base path where data files will be stored:

```go
factory := filesystem.NewFactory("/var/spool/retryspool/data")
```

## File Permissions

- Directories: `0755` (rwxr-xr-x)
- Files: Default OS permissions (usually `0644`)

## Error Handling

- Returns appropriate errors for missing files
- Handles filesystem errors gracefully
- Provides detailed error messages with context

## Cleanup

The backend automatically:
- Removes empty directories when deleting files
- Cleans up directory hierarchy recursively
- Preserves the base directory structure