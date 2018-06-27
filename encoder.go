// Copyright (c) 2014 José Carlos Nieto, https://menteslibres.net/xiam
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package resp

import (
	"io"
)

const digitbuflen = 20

func init()  {
	for i := 1; i < 128; i++{
		tiniIntBytesMap[i] = intToBytes(i)
	}
}

var (
	encoderNil      = []byte("$-1\r\n")
	digits          = []byte("0123456789")
	tiniIntBytesMap = make(map[int][]byte, 127)
	bytePoolMgr		= NewBytePoolManager()
)

func intToBytesInner(v int) []byte {
	buf := make([]byte, digitbuflen)

	i := len(buf)

	for v >= 10 {
		i--
		buf[i] = digits[v%10]
		v = v / 10
	}

	i--
	buf[i] = digits[v%10]

	return buf[i:]
}

func intToBytes(v int) []byte {
	if v < 128 {
		return tiniIntBytesMap[v]
	}
	return intToBytesInner(v)
}

// Encoder provides the Encode() method for encoding directly to an io.Writer.
type Encoder struct {
	w   io.Writer
	buf []byte
}

// NewEncoder creates and returns a *Encoder value with the given io.Writer.
func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{
		w:   w,
		buf: []byte{},
	}
	return e
}

// Encode marshals the given argument into a RESP message and pushes the output
// to the given writer.
func (e *Encoder) Encode(v interface{}) error {
	return e.writeEncoded(e.w, v)
}

func (e *Encoder) writeEncoded(w io.Writer, data interface{}) (err error) {

	var b []byte
	var bb []byte

	defer func() {
		if bb != nil{
			bytePoolMgr.Put(bb)
		}
	}()

	switch v := data.(type) {

	case []byte:
		n := intToBytes(len(v))

		bb = bytePoolMgr.Get(countCommndLength(data))

		b = append(bb, BulkHeader)
		b = append(b, n...)
		b = append(b, endOfLine...)
		b = append(b, v...)
		b = append(b, endOfLine...)

	case string:
		q := []byte(v)

		bb = bytePoolMgr.Get(countCommndLength(data))
		b = append(bb, StringHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

	case error:
		q := []byte(v.Error())

		bb = bytePoolMgr.Get(countCommndLength(data))
		b = append(bb, ErrorHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

	case int:
		q := intToBytes(int(v))
		bb = bytePoolMgr.Get(countCommndLength(data))
		b = append(bb, IntegerHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

	case [][]byte:
		n := intToBytes(len(v))

		bb = bytePoolMgr.Get(countCommndLength(data))
		b = append(bb, ArrayHeader)
		b = append(b, n...)
		b = append(b, endOfLine...)

		for i := range v {
			q := intToBytes(len(v[i]))

			b := append(b, BulkHeader)
			b = append(b, q...)
			b = append(b, endOfLine...)
			b = append(b, v[i]...)
			b = append(b, endOfLine...)
		}

	case []string:
		q := intToBytes(len(v))

		bb = bytePoolMgr.Get(countCommndLength(data))
		b = append(bb, ArrayHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

		for i := range v {
			p := []byte(v[i])

			b = append(b, StringHeader)
			b = append(b, p...)
			b = append(b, endOfLine...)
		}

	case []int:
		n := intToBytes(len(v))

		bb = bytePoolMgr.Get(countCommndLength(data))
		b = append(bb, ArrayHeader)
		b = append(b, n...)
		b = append(b, endOfLine...)

		for i := range v {
			m := intToBytes(v[i])

			b = append(b, IntegerHeader)
			b = append(b, m...)
			b = append(b, endOfLine...)
		}

	case []interface{}:
		q := intToBytes(len(v))
		bb = bytePoolMgr.Get(1+len(q)+2)
		b = append(bb, ArrayHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

		e.buf = append(e.buf, b...)

		if w != nil {
			w.Write(e.buf)
			e.buf = []byte{}
		}

		for i := range v {
			if err = e.writeEncoded(w, v[i]); err != nil {
				return err
			}
		}

		return nil

	case *Message:
		switch v.Type {
		case ErrorHeader:
			return e.writeEncoded(w, v.Error)
		case IntegerHeader:
			return e.writeEncoded(w, int(v.Integer))
		case BulkHeader:
			return e.writeEncoded(w, v.Bytes)
		case StringHeader:
			return e.writeEncoded(w, v.Status)
		case ArrayHeader:
			return e.writeEncoded(w, v.Array)
		default:
			return ErrMissingMessageHeader
		}

	case nil:
		b = encoderNil

	default:
		return ErrInvalidInput
	}

	e.buf = append(e.buf, b...)

	if w != nil {
		w.Write(e.buf)
		e.buf = []byte{}
	}

	return nil
}

func countCommndLength(data interface{}) int {
	switch v := data.(type) {

	case []byte:
		n := intToBytes(len(v))
		return 1+len(n)+2+len(v)+2

	case string:
		q := []byte(v)
		return 1+len(q)+2

	case error:
		q := []byte(v.Error())
		return 1+len(q)+2

	case int:
		q := intToBytes(int(v))
		return 1+len(q)+2

	case [][]byte:
		n := intToBytes(len(v))
		count := 1+len(n)+2
		for i := range v {
			q := intToBytes(len(v[i]))
			count += 1 + len(q) + 2 + len(v[i]) + 2
		}
		return count

	case []string:
		q := intToBytes(len(v))
		count :=1+len(q)+2
		for i := range v {
			p := []byte(v[i])
			count += 1 + len(p) + 2
		}
		return count

	case []int:
		n := intToBytes(len(v))
		count := 1+len(n)+2
		for i := range v {
			m := intToBytes(v[i])

			count += 1 + len(m) + 2
		}
		return count

	default:
		return -1
	}
}
