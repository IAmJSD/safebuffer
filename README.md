# safebuffer

A Go package providing a resizable buffer implementation with convenient methods for writing and reading various data types.

## Overview

The `safebuffer` package provides a `ResizableBuffer` type that allows for efficient buffer operations with automatic resizing. It supports writing various data types (integers, floats, strings, bytes) and includes both append and prepend operations.

## Method Chaining

All write operations return the buffer, allowing for convenient method chaining. Here are some examples:

```go
buf := NewResizableBuffer(nil).
    CopyString("Data: ").
    Uint32(42, true).
    CopyString(", ").
    Float64(3.14159, true).
    CRLF()
```

## Usage

### Creating a Buffer

```go
// Create an empty buffer
buf := NewResizableBuffer(nil)

// Create a buffer with a pre-made buffer
initialSlice := []byte{1, 2, 3}
buf := NewResizableBuffer(initialSlice)
```

### Writing Data

#### Basic Operations

```go
// Write bytes
buf.CopyBytes([]byte{1, 2, 3})

// Write string
buf.CopyString("Hello, World!")

// Write single byte
buf.Byte(0xFF)

// Write CRLF (carriage return + line feed)
buf.CRLF()
```

#### Integer Operations

```go
// Write uint16 (little endian)
buf.Uint16(1234, true)

// Write uint32 (big endian)
buf.Uint32(5678, false)

// Write uint64 (little endian)
buf.Uint64(9012, true)

// Write int16 (little endian)
buf.Int16(-1234, true)

// Write int32 (big endian)
buf.Int32(-5678, false)

// Write int64 (little endian)
buf.Int64(-9012, true)
```

#### Floating Point Operations

```go
// Write float32 (little endian)
buf.Float32(3.14, true)

// Write float64 (big endian)
buf.Float64(3.14159, false)
```

### Prepend Operations

```go
// Prepend bytes
buf.PrependBytes([]byte{1, 2, 3})

// Prepend string
buf.PrependString("Prefix")

// Prepend single byte
buf.PrependByte(0xFF)

// Prepend uint16 (little endian)
buf.PrependUint16(1234, true)

// Prepend uint32 (big endian)
buf.PrependUint32(5678, false)

// Prepend uint64 (little endian)
buf.PrependUint64(9012, true)

// Prepend int16 (little endian)
buf.PrependInt16(-1234, true)

// Prepend int32 (big endian)
buf.PrependInt32(-5678, false)

// Prepend int64 (little endian)
buf.PrependInt64(-9012, true)

// Prepend float32 (little endian)
buf.PrependFloat32(3.14, true)

// Prepend float64 (big endian)
buf.PrependFloat64(3.14159, false)
```

### Reading and Management

```go
// Get the current buffer contents
data := buf.Bytes()

// Get the current length
length := buf.Len()

// Reset the buffer (optionally zero out the contents)
buf.Reset(true) // true to zero out, false to keep contents

// Read from an io.Reader into the buffer
chunk, err := buf.ReadInto(reader, maxSize)

// Create a sub-buffer
subBuf := buf.SubBuffer(length) // length < 0 for remaining buffer
```

## Function Reference

### Constructor
- `NewResizableBuffer(b []byte) *ResizableBuffer` - Creates a new resizable buffer, optionally with initial data

### Basic Operations
- `CopyBytes(p []byte) *ResizableBuffer` - Copies bytes into the buffer
- `CopyString(p string) *ResizableBuffer` - Copies a string into the buffer
- `Byte(bt byte) *ResizableBuffer` - Writes a single byte
- `CRLF() *ResizableBuffer` - Writes CRLF (carriage return + line feed)

### Integer Operations
- `Uint16(v uint16, littleEndian bool) *ResizableBuffer` - Writes uint16
- `Uint32(v uint32, littleEndian bool) *ResizableBuffer` - Writes uint32
- `Uint64(v uint64, littleEndian bool) *ResizableBuffer` - Writes uint64
- `Int16(v int16, littleEndian bool) *ResizableBuffer` - Writes int16
- `Int32(v int32, littleEndian bool) *ResizableBuffer` - Writes int32
- `Int64(v int64, littleEndian bool) *ResizableBuffer` - Writes int64

### Floating Point Operations
- `Float32(v float32, littleEndian bool) *ResizableBuffer` - Writes float32
- `Float64(v float64, littleEndian bool) *ResizableBuffer` - Writes float64

### Prepend Operations
- `PrependBytes(v []byte) *ResizableBuffer` - Prepends bytes
- `PrependString(v string) *ResizableBuffer` - Prepends a string
- `PrependByte(v byte) *ResizableBuffer` - Prepends a byte
- `PrependUint16(v uint16, littleEndian bool) *ResizableBuffer` - Prepends uint16
- `PrependUint32(v uint32, littleEndian bool) *ResizableBuffer` - Prepends uint32
- `PrependUint64(v uint64, littleEndian bool) *ResizableBuffer` - Prepends uint64
- `PrependInt16(v int16, littleEndian bool) *ResizableBuffer` - Prepends int16
- `PrependInt32(v int32, littleEndian bool) *ResizableBuffer` - Prepends int32
- `PrependInt64(v int64, littleEndian bool) *ResizableBuffer` - Prepends int64
- `PrependFloat32(v float32, littleEndian bool) *ResizableBuffer` - Prepends float32
- `PrependFloat64(v float64, littleEndian bool) *ResizableBuffer` - Prepends float64

### Buffer Management
- `Bytes() []byte` - Returns the current buffer contents
- `Len() int` - Returns the current buffer length
- `Reset(zeroOut bool) *ResizableBuffer` - Resets the buffer, optionally zeroing out contents
- `ReadInto(r io.Reader, maxSize int) ([]byte, error)` - Reads from an io.Reader into the buffer
- `SubBuffer(length int) *ResizableBuffer` - Creates a sub-buffer view of the current buffer

## Notes

- The buffer automatically resizes when needed
- All write operations return the buffer for method chaining
- The buffer is not thread-safe
- Bytes returned by `Bytes()` and `ReadInto()` are only valid until the next `Reset()` call
- `SubBuffer()` creates a view into the current buffer, not a copy
