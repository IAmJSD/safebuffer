package safebuffer

import (
	"encoding/binary"
	"io"
	"math"
)

// ResizableBuffer is a buffer that can be resized. This is single threaded.
type ResizableBuffer struct {
	buffer []byte
	offset int
}

// NewResizableBuffer creates a new ResizableBuffer. Can be nil if you want it to be fully dynamic.
func NewResizableBuffer(b []byte) *ResizableBuffer {
	if b == nil {
		b = []byte{}
	}
	return &ResizableBuffer{
		buffer: b,
	}
}

func (b *ResizableBuffer) ensureCapacity(n int) {
	if len(b.buffer)-b.offset < n {
		buf := make([]byte, n+len(b.buffer))
		copy(buf, b.buffer)
		b.buffer = buf
	}
}

// CopyBytes copies the bytes specified into the consumed buffer.
func (b *ResizableBuffer) CopyBytes(p []byte) *ResizableBuffer {
	b.ensureCapacity(len(p))
	copy(b.buffer[b.offset:], p)
	b.offset += len(p)
	return b
}

// CopyString copies the string specified into the consumed buffer.
func (b *ResizableBuffer) CopyString(p string) *ResizableBuffer {
	b.ensureCapacity(len(p))
	copy(b.buffer[b.offset:], p)
	b.offset += len(p)
	return b
}

// Byte writes a single byte into the consumed buffer.
func (b *ResizableBuffer) Byte(bt byte) *ResizableBuffer {
	b.ensureCapacity(1)
	b.buffer[b.offset] = bt
	b.offset++
	return b
}

// CRLF writes a CRLF into the consumed buffer.
func (b *ResizableBuffer) CRLF() *ResizableBuffer {
	b.ensureCapacity(2)
	b.buffer[b.offset] = '\r'
	b.buffer[b.offset+1] = '\n'
	b.offset += 2
	return b
}

// Uint16 writes a uint16 into the consumed buffer.
func (b *ResizableBuffer) Uint16(v uint16, littleEndian bool) *ResizableBuffer {
	b.ensureCapacity(2)
	if littleEndian {
		binary.LittleEndian.PutUint16(b.buffer[b.offset:], v)
	} else {
		binary.BigEndian.PutUint16(b.buffer[b.offset:], v)
	}
	b.offset += 2
	return b
}

// Uint32 writes a uint32 into the consumed buffer.
func (b *ResizableBuffer) Uint32(v uint32, littleEndian bool) *ResizableBuffer {
	b.ensureCapacity(4)
	if littleEndian {
		binary.LittleEndian.PutUint32(b.buffer[b.offset:], v)
	} else {
		binary.BigEndian.PutUint32(b.buffer[b.offset:], v)
	}
	b.offset += 4
	return b
}

// Uint64 writes a uint64 into the consumed buffer.
func (b *ResizableBuffer) Uint64(v uint64, littleEndian bool) *ResizableBuffer {
	b.ensureCapacity(8)
	if littleEndian {
		binary.LittleEndian.PutUint64(b.buffer[b.offset:], v)
	} else {
		binary.BigEndian.PutUint64(b.buffer[b.offset:], v)
	}
	b.offset += 8
	return b
}

// Int16 writes a int16 into the consumed buffer.
func (b *ResizableBuffer) Int16(v int16, littleEndian bool) *ResizableBuffer {
	return b.Uint16(uint16(v), littleEndian)
}

// Int32 writes a int32 into the consumed buffer.
func (b *ResizableBuffer) Int32(v int32, littleEndian bool) *ResizableBuffer {
	return b.Uint32(uint32(v), littleEndian)
}

// Int64 writes a int64 into the consumed buffer.
func (b *ResizableBuffer) Int64(v int64, littleEndian bool) *ResizableBuffer {
	return b.Uint64(uint64(v), littleEndian)
}

// Float32 writes a float32 into the consumed buffer.
func (b *ResizableBuffer) Float32(v float32, littleEndian bool) *ResizableBuffer {
	return b.Uint32(math.Float32bits(v), littleEndian)
}

// Float64 writes a float64 into the consumed buffer.
func (b *ResizableBuffer) Float64(v float64, littleEndian bool) *ResizableBuffer {
	return b.Uint64(math.Float64bits(v), littleEndian)
}

// Bytes returns the bytes of the consumed buffer. Note that this slice is only valid
// until the next call to Reset. After that function is called, it is not guaranteed that
// this data will not be overwritten.
func (b *ResizableBuffer) Bytes() []byte {
	return b.buffer[:b.offset]
}

// Len returns the length of the consumed buffer.
func (b *ResizableBuffer) Len() int {
	return b.offset
}

func (b *ResizableBuffer) prependStart(n int, f func(b []byte)) *ResizableBuffer {
	if len(b.buffer)-b.offset < n {
		buf := make([]byte, n+len(b.buffer), n+len(b.buffer))
		copy(buf[n:], b.buffer)
		b.buffer = buf
	} else {
		copy(b.buffer[n:], b.buffer)
	}
	f(b.buffer)
	b.offset += n
	return b
}

// PrependBytes prepends a byte slice into the consumed buffer.
func (b *ResizableBuffer) PrependBytes(v []byte) *ResizableBuffer {
	return b.prependStart(len(v), func(b []byte) {
		copy(b, v)
	})
}

// PrependByte prepends a byte into the consumed buffer.
func (b *ResizableBuffer) PrependByte(v byte) *ResizableBuffer {
	return b.prependStart(1, func(b []byte) {
		b[0] = v
	})
}

// PrependString prepends a string into the consumed buffer.
func (b *ResizableBuffer) PrependString(v string) *ResizableBuffer {
	return b.prependStart(len(v), func(b []byte) {
		copy(b, v)
	})
}

// PrependUint16 prepends a uint16 into the consumed buffer.
func (b *ResizableBuffer) PrependUint16(v uint16, littleEndian bool) *ResizableBuffer {
	return b.prependStart(2, func(b []byte) {
		if littleEndian {
			binary.LittleEndian.PutUint16(b, v)
		} else {
			binary.BigEndian.PutUint16(b, v)
		}
	})
}

// PrependUint32 prepends a uint32 into the consumed buffer.
func (b *ResizableBuffer) PrependUint32(v uint32, littleEndian bool) *ResizableBuffer {
	return b.prependStart(4, func(b []byte) {
		if littleEndian {
			binary.LittleEndian.PutUint32(b, v)
		} else {
			binary.BigEndian.PutUint32(b, v)
		}
	})
}

// PrependUint64 prepends a uint64 into the consumed buffer.
func (b *ResizableBuffer) PrependUint64(v uint64, littleEndian bool) *ResizableBuffer {
	return b.prependStart(8, func(b []byte) {
		if littleEndian {
			binary.LittleEndian.PutUint64(b, v)
		} else {
			binary.BigEndian.PutUint64(b, v)
		}
	})
}

// PrependInt16 prepends a int16 into the consumed buffer.
func (b *ResizableBuffer) PrependInt16(v int16, littleEndian bool) *ResizableBuffer {
	return b.PrependUint16(uint16(v), littleEndian)
}

// PrependInt32 prepends a int32 into the consumed buffer.
func (b *ResizableBuffer) PrependInt32(v int32, littleEndian bool) *ResizableBuffer {
	return b.PrependUint32(uint32(v), littleEndian)
}

// PrependInt64 prepends a int64 into the consumed buffer.
func (b *ResizableBuffer) PrependInt64(v int64, littleEndian bool) *ResizableBuffer {
	return b.PrependUint64(uint64(v), littleEndian)
}

// PrependFloat32 prepends a float32 into the consumed buffer.
func (b *ResizableBuffer) PrependFloat32(v float32, littleEndian bool) *ResizableBuffer {
	return b.PrependUint32(math.Float32bits(v), littleEndian)
}

// PrependFloat64 prepends a float64 into the consumed buffer.
func (b *ResizableBuffer) PrependFloat64(v float64, littleEndian bool) *ResizableBuffer {
	return b.PrependUint64(math.Float64bits(v), littleEndian)
}

// Reset resets the consumed buffer so it can be reused. Can optionally zero out
// the buffer data we wrote to prevent information leaks.
func (b *ResizableBuffer) Reset(zeroOut bool) *ResizableBuffer {
	if zeroOut {
		clear(b.buffer[:b.offset])
	}
	b.offset = 0
	return b
}

// ReadInto is used to read into a buffer from a io.Reader. The returned slice is only valid
// until the next call to Reset. After that function is called, it is not guaranteed that
// this data will not be overwritten.
func (b *ResizableBuffer) ReadInto(r io.Reader, maxSize int) ([]byte, error) {
	if maxSize > 0 {
		b.ensureCapacity(maxSize)
	} else {
		// The maximum size is the length of the buffer minus the offset
		maxSize = len(b.buffer) - b.offset
	}

	n, err := r.Read(b.buffer[b.offset : b.offset+maxSize])
	chunk := b.buffer[b.offset : b.offset+n]
	b.offset += n
	return chunk, err
}

// SubBuffer returns a new ResizableBuffer that is a subbuffer of the current one.
// If you specify <0 for the length, it will use the remaining buffer. The returned
// buffer is not a copy, it is a view into the current buffer. Note that this means
// that the returned buffer is only valid until the next call to Reset. After that
// function is called, it is not guaranteed that the returned buffer will not be
// overwritten.
func (b *ResizableBuffer) SubBuffer(length int) *ResizableBuffer {
	if length < 0 {
		length = len(b.buffer) - b.offset
	}
	b.ensureCapacity(length)
	s := b.buffer[b.offset : b.offset+length]
	b.offset += length
	return &ResizableBuffer{buffer: s}
}
