package safebuffer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func TestNewResizableBuffer(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		rb := NewResizableBuffer(nil)
		if rb == nil {
			t.Fatal("NewResizableBuffer(nil) should return a non-nil ResizableBuffer")
		}
		if rb.buffer == nil {
			t.Fatal("NewResizableBuffer(nil) should return a ResizableBuffer without a nil buffer")
		}
		if len(rb.buffer) != 0 {
			t.Fatal("NewResizableBuffer(nil) should return a ResizableBuffer with a buffer of length 0")
		}
	})

	t.Run("buffer given", func(t *testing.T) {
		s := []byte{1, 2, 3}
		rb := NewResizableBuffer(s)
		if rb == nil {
			t.Fatal("NewResizableBuffer(nil) should return a non-nil ResizableBuffer")
		}
		if len(rb.buffer) != len(s) {
			t.Fatalf("NewResizableBuffer(s) should return a ResizableBuffer with the same buffer as s, got %v", rb.buffer)
		}
		if &rb.buffer[0] != &s[0] {
			t.Fatal("NewResizableBuffer(s) should return a ResizableBuffer with the same buffer as s")
		}
	})
}

type testCase struct {
	name string
	eq   []byte
	fn   func(t *testing.T, b *ResizableBuffer)
}

func handleChainCase(fn func(b *ResizableBuffer) *ResizableBuffer) func(t *testing.T, b *ResizableBuffer) {
	return func(t *testing.T, b *ResizableBuffer) {
		x := fn(b)
		if x != b {
			t.Fatal("expected fn to return the buffer")
		}
	}
}

func sharedBufferTestCases(t *testing.T, firstCase testCase) {
	t.Run("empty buffer", func(t *testing.T) {
		rb := NewResizableBuffer(nil)
		firstCase.fn(t, rb)
		if rb.offset != len(firstCase.eq) {
			t.Fatalf("expected offset to be %d, got %d", len(firstCase.eq), rb.offset)
		}
		b := rb.Bytes()
		if !bytes.Equal(b, firstCase.eq) {
			t.Fatalf("expected %v, got %v", firstCase.eq, b)
		}
	})
	t.Run("large buffer", func(t *testing.T) {
		buffer := make([]byte, 1000)
		rb := NewResizableBuffer(buffer)
		firstCase.fn(t, rb)
		if &rb.buffer[0] != &buffer[0] {
			t.Fatal("expected buffer to be un-changed")
		}
		if rb.offset != len(firstCase.eq) {
			t.Fatalf("expected offset to be %d, got %d", len(firstCase.eq), rb.offset)
		}
		b := rb.Bytes()
		if !bytes.Equal(b, firstCase.eq) {
			t.Fatalf("expected %v, got %v", firstCase.eq, b)
		}
	})
}

func createMainTestCase(t *testing.T, test testCase, expected []byte, bigBufferOnly bool) {
	initBuffer := []byte{1}
	if bigBufferOnly {
		initBuffer = make([]byte, 1000)
		initBuffer[0] = 1
	}
	rb := NewResizableBuffer(initBuffer)
	rb.offset = 1
	test.fn(t, rb)
	if rb.offset != len(test.eq)+1 {
		t.Fatalf("expected offset to be %d, got %d", len(test.eq)+1, rb.offset)
	}
	b := rb.Bytes()
	if bigBufferOnly {
		if &rb.buffer[0] != &initBuffer[0] {
			t.Fatal("expected buffer to not be swapped")
		}
	} else {
		if len(test.eq) > 1 && &rb.buffer[0] == &initBuffer[0] {
			t.Fatal("expected buffer to be swapped")
		}
	}
	if !bytes.Equal(b, expected) {
		t.Fatalf("expected %v, got %v", expected, b)
	}
}

func testAppendCases(t *testing.T, tests []testCase, bigBufferOnly bool) {
	if !bigBufferOnly {
		sharedBufferTestCases(t, tests[0])
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := make([]byte, len(test.eq)+1)
			expected[0] = 1
			copy(expected[1:], test.eq)
			createMainTestCase(t, test, expected, bigBufferOnly)
		})
	}
}

func TestCopyBytes(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "content",
			eq:   []byte{1, 2, 3},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.CopyBytes([]byte{1, 2, 3})
			}),
		},
		{
			name: "empty",
			eq:   []byte{},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.CopyBytes([]byte{})
			}),
		},
		{
			name: "nil",
			eq:   []byte{},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.CopyBytes(nil)
			}),
		},
	}, false)
}

func TestCopyString(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "content",
			eq:   []byte("content"),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.CopyString("content")
			}),
		},
		{
			name: "empty",
			eq:   []byte(""),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.CopyString("")
			}),
		},
	}, false)
}

func TestByte(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "content",
			eq:   []byte{1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Byte(1)
			}),
		},
		{
			name: "null",
			eq:   []byte{0},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Byte(0)
			}),
		},
	}, false)
}

func TestCRLF(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "writes",
			eq:   []byte("\r\n"),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.CRLF()
			}),
		},
	}, false)
}

func TestUint16(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian",
			eq:   []byte{2, 1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Uint16(0x0102, true)
			}),
		},
		{
			name: "big endian",
			eq:   []byte{1, 2},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Uint16(0x0102, false)
			}),
		},
	}, false)
}

func TestUint32(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian",
			eq:   []byte{4, 3, 2, 1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Uint32(0x01020304, true)
			}),
		},
		{
			name: "big endian",
			eq:   []byte{1, 2, 3, 4},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Uint32(0x01020304, false)
			}),
		},
	}, false)
}

func TestUint64(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian",
			eq:   []byte{8, 7, 6, 5, 4, 3, 2, 1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Uint64(0x0102030405060708, true)
			}),
		},
		{
			name: "big endian",
			eq:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Uint64(0x0102030405060708, false)
			}),
		},
	}, false)
}

func slowButKnownGoodValConv(a any, littleEndian bool) []byte {
	var buf bytes.Buffer
	if littleEndian {
		binary.Write(&buf, binary.LittleEndian, a)
	} else {
		binary.Write(&buf, binary.BigEndian, a)
	}
	return buf.Bytes()
}

func TestInt16(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian negative",
			eq:   slowButKnownGoodValConv(int16(-1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int16(-1, true)
			}),
		},
		{
			name: "big endian negative",
			eq:   slowButKnownGoodValConv(int16(-1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int16(-1, false)
			}),
		},
		{
			name: "little endian positive",
			eq:   slowButKnownGoodValConv(int16(1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int16(1, true)
			}),
		},
		{
			name: "big endian positive",
			eq:   slowButKnownGoodValConv(int16(1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int16(1, false)
			}),
		},
	}, false)
}

func TestInt32(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian negative",
			eq:   slowButKnownGoodValConv(int32(-1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int32(-1, true)
			}),
		},
		{
			name: "big endian negative",
			eq:   slowButKnownGoodValConv(int32(-1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int32(-1, false)
			}),
		},
		{
			name: "little endian positive",
			eq:   slowButKnownGoodValConv(int32(1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int32(1, true)
			}),
		},
		{
			name: "big endian positive",
			eq:   slowButKnownGoodValConv(int32(1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int32(1, false)
			}),
		},
	}, false)
}

func TestInt64(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian negative",
			eq:   slowButKnownGoodValConv(int64(-1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int64(-1, true)
			}),
		},
		{
			name: "big endian negative",
			eq:   slowButKnownGoodValConv(int64(-1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int64(-1, false)
			}),
		},
		{
			name: "little endian positive",
			eq:   slowButKnownGoodValConv(int64(1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int64(1, true)
			}),
		},
		{
			name: "big endian positive",
			eq:   slowButKnownGoodValConv(int64(1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Int64(1, false)
			}),
		},
	}, false)
}

func TestFloat32(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian",
			eq:   slowButKnownGoodValConv(float32(1.2), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Float32(1.2, true)
			}),
		},
		{
			name: "big endian",
			eq:   slowButKnownGoodValConv(float32(1.2), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Float32(1.2, false)
			}),
		},
	}, false)
}

func TestFloat64(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "little endian",
			eq:   slowButKnownGoodValConv(float64(1.2), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Float64(1.2, true)
			}),
		},
		{
			name: "big endian",
			eq:   slowButKnownGoodValConv(float64(1.2), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.Float64(1.2, false)
			}),
		},
	}, false)
}

func TestBytes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		rb := NewResizableBuffer(nil)
		b := rb.Bytes()
		if len(b) != 0 {
			t.Fatalf("expected 0 bytes, got %d", len(b))
		}
		rb.CopyBytes([]byte{1, 2, 3})
		b = rb.Bytes()
		if len(b) != 3 {
			t.Fatalf("expected 3 bytes, got %d", len(b))
		}
		if &b[0] != &rb.buffer[0] {
			t.Fatalf("expected buffer to be un-changed")
		}
	})

	t.Run("large buffer", func(t *testing.T) {
		buffer := make([]byte, 1000)
		rb := NewResizableBuffer(buffer)
		b := rb.Bytes()
		if len(b) != 0 {
			t.Fatalf("expected 0 bytes, got %d", len(b))
		}
		rb.CopyBytes([]byte{1, 2, 3})
		b = rb.Bytes()
		if len(b) != 3 {
			t.Fatalf("expected 3 bytes, got %d", len(b))
		}
		if &b[0] != &rb.buffer[0] && &b[0] != &buffer[0] {
			t.Fatalf("expected buffer to be un-changed")
		}
	})
}

func TestLen(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		rb := NewResizableBuffer(nil)
		if rb.Len() != 0 {
			t.Fatalf("expected 0 bytes, got %d", rb.Len())
		}
		rb.Byte(' ')
		if rb.Len() != 1 {
			t.Fatalf("expected 1 byte, got %d", rb.Len())
		}
	})

	t.Run("large buffer", func(t *testing.T) {
		buffer := make([]byte, 1000)
		rb := NewResizableBuffer(buffer)
		if rb.Len() != 0 {
			t.Fatalf("expected 0 bytes, got %d", rb.Len())
		}
		rb.Byte(' ')
		if rb.Len() != 1 {
			t.Fatalf("expected 1 byte, got %d", rb.Len())
		}
	})
}

func testPrependCases(t *testing.T, tests []testCase) {
	sharedBufferTestCases(t, tests[0])

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := make([]byte, len(test.eq)+1)
			copy(expected, test.eq)
			expected[len(test.eq)] = 1
			createMainTestCase(t, test, expected, false)
		})
	}
}

func TestPrependBytes(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "content",
			eq:   []byte{1, 2, 3},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependBytes([]byte{1, 2, 3})
			}),
		},
		{
			name: "empty",
			eq:   []byte{},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependBytes([]byte{})
			}),
		},
		{
			name: "nil",
			eq:   []byte{},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependBytes(nil)
			}),
		},
	})
}

func TestPrependByte(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "content",
			eq:   []byte{1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependByte(1)
			}),
		},
		{
			name: "null",
			eq:   []byte{0},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependByte(0)
			}),
		},
	})
}

func TestPrependString(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "content",
			eq:   []byte("content"),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependString("content")
			}),
		},
		{
			name: "empty",
			eq:   []byte(""),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependString("")
			}),
		},
	})
}

func TestPrependUint16(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian",
			eq:   []byte{2, 1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependUint16(0x0102, true)
			}),
		},
		{
			name: "big endian",
			eq:   []byte{1, 2},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependUint16(0x0102, false)
			}),
		},
	})
}

func TestPrependUint32(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian",
			eq:   []byte{4, 3, 2, 1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependUint32(0x01020304, true)
			}),
		},
		{
			name: "big endian",
			eq:   []byte{1, 2, 3, 4},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependUint32(0x01020304, false)
			}),
		},
	})
}

func TestPrependUint64(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian",
			eq:   []byte{8, 7, 6, 5, 4, 3, 2, 1},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependUint64(0x0102030405060708, true)
			}),
		},
		{
			name: "big endian",
			eq:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependUint64(0x0102030405060708, false)
			}),
		},
	})
}

func TestPrependInt16(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian negative",
			eq:   slowButKnownGoodValConv(int16(-1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt16(-1, true)
			}),
		},
		{
			name: "big endian negative",
			eq:   slowButKnownGoodValConv(int16(-1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt16(-1, false)
			}),
		},
		{
			name: "little endian positive",
			eq:   slowButKnownGoodValConv(int16(1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt16(1, true)
			}),
		},
		{
			name: "big endian positive",
			eq:   slowButKnownGoodValConv(int16(1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt16(1, false)
			}),
		},
	})
}

func TestPrependInt32(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian negative",
			eq:   slowButKnownGoodValConv(int32(-1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt32(-1, true)
			}),
		},
		{
			name: "big endian negative",
			eq:   slowButKnownGoodValConv(int32(-1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt32(-1, false)
			}),
		},
		{
			name: "little endian positive",
			eq:   slowButKnownGoodValConv(int32(1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt32(1, true)
			}),
		},
		{
			name: "big endian positive",
			eq:   slowButKnownGoodValConv(int32(1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt32(1, false)
			}),
		},
	})
}

func TestPrependInt64(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian negative",
			eq:   slowButKnownGoodValConv(int64(-1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt64(-1, true)
			}),
		},
		{
			name: "big endian negative",
			eq:   slowButKnownGoodValConv(int64(-1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt64(-1, false)
			}),
		},
		{
			name: "little endian positive",
			eq:   slowButKnownGoodValConv(int64(1), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt64(1, true)
			}),
		},
		{
			name: "big endian positive",
			eq:   slowButKnownGoodValConv(int64(1), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependInt64(1, false)
			}),
		},
	})
}

func TestPrependFloat32(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian",
			eq:   slowButKnownGoodValConv(float32(1.2), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependFloat32(1.2, true)
			}),
		},
		{
			name: "big endian",
			eq:   slowButKnownGoodValConv(float32(1.2), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependFloat32(1.2, false)
			}),
		},
	})
}

func TestPrependFloat64(t *testing.T) {
	testPrependCases(t, []testCase{
		{
			name: "little endian",
			eq:   slowButKnownGoodValConv(float64(1.2), true),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependFloat64(1.2, true)
			}),
		},
		{
			name: "big endian",
			eq:   slowButKnownGoodValConv(float64(1.2), false),
			fn: handleChainCase(func(b *ResizableBuffer) *ResizableBuffer {
				return b.PrependFloat64(1.2, false)
			}),
		},
	})
}

func TestReset(t *testing.T) {
	buildAndUse := func(name string, fn func(*ResizableBuffer)) {
		t.Run(name, func(t *testing.T) {
			rb := NewResizableBuffer(make([]byte, 1000))
			rb.CopyString("hello")
			rb.Uint16(0x0102, true)
			rb.Uint32(0x01020304, false)
			fn(rb)
		})
	}
	buildAndUse("zero out", func(rb *ResizableBuffer) {
		underlyingData := rb.buffer
		underlyingData[100] = 1
		limited := underlyingData[:rb.offset]
		rb.Reset(true)
		if underlyingData[100] != 1 {
			t.Fatal("expected data outside of usage not to be zeroed out")
		}
		if rb.Len() != 0 {
			t.Fatalf("expected 0 bytes, got %d", rb.Len())
		}
		for _, b := range limited {
			if b != 0 {
				t.Fatal("expected data to be zeroed out")
			}
		}
	})
	buildAndUse("no zero out", func(rb *ResizableBuffer) {
		underlyingData := rb.buffer[:rb.offset]
		rb.Reset(false)
		for _, b := range underlyingData {
			if b == 0 {
				t.Fatal("expected data to not be zeroed out")
			}
		}
	})
}

type stringReader struct {
	s string
}

func (r stringReader) Read(p []byte) (n int, err error) {
	return copy(p, r.s), nil
}

type errorfulReader struct {
	err error
}

func (r errorfulReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func testReaderCase(r io.Reader, maxSize int) func(t *testing.T, b *ResizableBuffer) {
	return func(t *testing.T, b *ResizableBuffer) {
		var errExpected error
		var expectedBytes []byte
		switch x := r.(type) {
		case stringReader:
			expectedBytes = []byte(x.s)
		case errorfulReader:
			errExpected = x.err
		default:
			t.Fatalf("unexpected reader type: %T", r)
		}
		by, err := b.ReadInto(r, maxSize)
		if err != errExpected {
			if err == nil {
				t.Fatalf("expected error %v, got nil", errExpected)
			}
			if errExpected == nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		}
		reslice := expectedBytes
		if maxSize > 0 {
			if maxSize > len(reslice) {
				maxSize = len(reslice)
			}
			reslice = reslice[:maxSize]
		}
		if !bytes.Equal(by, reslice) {
			t.Fatalf("expected %v, got %v", reslice, by)
		}
	}
}

func TestReadInto(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "unbound string reader",
			eq:   []byte("hello"),
			fn:   testReaderCase(stringReader{s: "hello"}, -1),
		},
		{
			name: "bound string reader full read",
			eq:   []byte("hello"),
			fn:   testReaderCase(stringReader{s: "hello"}, 5),
		},
		{
			name: "bound string reader partial read",
			eq:   []byte("hell"),
			fn:   testReaderCase(stringReader{s: "hello"}, 4),
		},
		{
			name: "bound larger than buffer",
			eq:   []byte("hello"),
			fn:   testReaderCase(stringReader{s: "hello"}, 100),
		},
		{
			name: "empty reader",
			eq:   []byte{},
			fn:   testReaderCase(stringReader{s: ""}, -1),
		},
		{
			name: "reader errors",
			eq:   []byte{},
			fn:   testReaderCase(errorfulReader{err: errors.New("test error")}, -1),
		},
	}, true)
}

func testSubBufferCase(bufLen int) func(t *testing.T, b *ResizableBuffer) {
	return func(t *testing.T, b *ResizableBuffer) {
		b.Byte('A')
		expectedLen := bufLen
		if expectedLen < 0 {
			expectedLen = len(b.buffer) - b.offset
		}

		sub := b.SubBuffer(bufLen)
		if len(sub.buffer) != expectedLen {
			t.Fatalf("expected %d bytes, got %d", expectedLen, len(sub.buffer))
		}
		if sub.offset != 0 {
			t.Fatalf("expected offset to be 0, got %d", sub.offset)
		}
		sub.CopyString("BCD")
	}
}

func abcdPadded() []byte {
	s := make([]byte, 999)
	copy(s, "ABCD")
	return s
}

func TestSubBuffer(t *testing.T) {
	testAppendCases(t, []testCase{
		{
			name: "unbound subbuffer",
			eq:   abcdPadded(),
			fn:   testSubBufferCase(-1),
		},
		{
			name: "bound subbuffer",
			eq:   []byte("ABCD"),
			fn:   testSubBufferCase(3),
		},
	}, true)
}
