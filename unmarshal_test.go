package jsonstream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/tada/catch"
	"github.com/tada/catch/pio"
)

func decoderOn(s string) Decoder {
	return NewDecoder(bytes.NewReader([]byte(s)))
}

type ts struct {
	v time.Duration
}

func (t *ts) MarshalJSON() ([]byte, error) {
	return Marshal(t)
}

func (t *ts) UnmarshalJSON(bs []byte) error {
	return Unmarshal(t, bs)
}

func (t *ts) MarshalToJSON(w io.Writer) {
	pio.WriteByte('{', w)
	WriteString("v", w)
	pio.WriteByte(':', w)
	pio.WriteInt(int64(t.v/time.Millisecond), w)
	pio.WriteByte('}', w)
}

func (t *ts) UnmarshalFromJSON(js Decoder, firstToken json.Token) {
	AssertDelim(firstToken, '{')
	for {
		s, ok := js.ReadStringOrEnd('}')
		if !ok {
			break
		}
		if s == "v" {
			t.v = time.Duration(js.ReadInt()) * time.Millisecond
		}
	}
}

func ExampleUnmarshal() {
	tv := ts{}
	err := json.Unmarshal([]byte(`{"v":38}`), &tv)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&tv)
	// Output: &{38000000}
}

func TestReadDelim(t *testing.T) {
	js := decoderOn("{}")
	err := catch.Do(func() {
		js.ReadDelim('{')
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		js.ReadDelim('{')
	})
	if err == nil {
		t.Fatal("expected error")
	}
	err = catch.Do(func() {
		js.ReadDelim('{')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadString(t *testing.T) {
	js := decoderOn(`"a"`)
	err := catch.Do(func() {
		if s := js.ReadString(); s != "a" {
			panic(catch.Error(`expected "a", got "%s"`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`1`)
	err = catch.Do(func() {
		js.ReadString()
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`null`)
	err = catch.Do(func() {
		if s := js.ReadString(); s != "" {
			panic(catch.Error(`expected empty string, got %q`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadStringOrEnd(t *testing.T) {
	js := decoderOn(`["a", null]`)
	err := catch.Do(func() {
		js.ReadDelim('[')
		if s, ok := js.ReadStringOrEnd(']'); !(ok && s == "a") {
			panic(catch.Error(`expected "a", got end`))
		}
		if s, ok := js.ReadStringOrEnd(']'); !(ok && s == "") {
			panic(catch.Error(`expected null, got end`))
		}
		if s, ok := js.ReadStringOrEnd(']'); ok {
			panic(catch.Error(`expected end, got "%s"`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = js.ReadStringOrEnd(']')
	})
	if err != io.ErrUnexpectedEOF {
		t.Fatal("expected ErrUnexpectedEOF")
	}
	js = decoderOn(`23`)
	err = catch.Do(func() {
		_, _ = js.ReadStringOrEnd(']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadBool(t *testing.T) {
	js := decoderOn(`true`)
	err := catch.Do(func() {
		if s := js.ReadBool(); !s {
			panic(catch.Error(`expected true, got false`))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		js.ReadBool()
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`null`)
	err = catch.Do(func() {
		if s := js.ReadBool(); s {
			panic(catch.Error(`expected false, got true`))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadBoolOrEnd(t *testing.T) {
	js := decoderOn(`[true, false, null]`)
	err := catch.Do(func() {
		js.ReadDelim('[')
		if b, ok := js.ReadBoolOrEnd(']'); !(ok && b) {
			panic(catch.Error(`expected true, got end`))
		}
		if b, ok := js.ReadBoolOrEnd(']'); !(ok && !b) {
			panic(catch.Error(`expected false, got end`))
		}
		if b, ok := js.ReadBoolOrEnd(']'); !(ok && !b) {
			panic(catch.Error(`expected false, got end`))
		}
		if b, ok := js.ReadBoolOrEnd(']'); ok {
			panic(catch.Error(`expected end, got "%b"`, b))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = js.ReadBoolOrEnd(']')
	})
	if err != io.ErrUnexpectedEOF {
		t.Fatal("expected ErrUnexpectedEOF")
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		_, _ = js.ReadBoolOrEnd(']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadInt(t *testing.T) {
	js := decoderOn(`1`)
	err := catch.Do(func() {
		if s := js.ReadInt(); s != 1 {
			panic(catch.Error(`expected 1, got %d`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		js.ReadInt()
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`null`)
	err = catch.Do(func() {
		if s := js.ReadInt(); s != 0 {
			panic(catch.Error(`expected 0, got %d`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadIntOrEnd(t *testing.T) {
	js := decoderOn(`[42, null]`)
	err := catch.Do(func() {
		js.ReadDelim('[')
		if i, ok := js.ReadIntOrEnd(']'); !(ok && i == 42) {
			panic(catch.Error(`expected 42, got end`))
		}
		if i, ok := js.ReadIntOrEnd(']'); !(ok && i == 0) {
			panic(catch.Error(`expected 0, got end`))
		}
		if i, ok := js.ReadIntOrEnd(']'); ok {
			panic(catch.Error(`expected end, got "%d"`, i))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = js.ReadIntOrEnd(']')
	})
	if err != io.ErrUnexpectedEOF {
		t.Fatal("expected ErrUnexpectedEOF")
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		_, _ = js.ReadIntOrEnd(']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadfloat(t *testing.T) {
	js := decoderOn(`1.3`)
	err := catch.Do(func() {
		if s := js.ReadFloat(); s != 1.3 {
			panic(catch.Error(`expected 1.3, got %f`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		js.ReadFloat()
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`null`)
	err = catch.Do(func() {
		if s := js.ReadFloat(); s != 0 {
			panic(catch.Error(`expected 0, got %f`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadFloatOrEnd(t *testing.T) {
	js := decoderOn(`[42.31, null]`)
	err := catch.Do(func() {
		js.ReadDelim('[')
		if f, ok := js.ReadFloatOrEnd(']'); !(ok && f == 42.31) {
			panic(catch.Error(`expected 42, got end`))
		}
		if f, ok := js.ReadFloatOrEnd(']'); !(ok && f == 0) {
			panic(catch.Error(`expected 0, got end`))
		}
		if f, ok := js.ReadFloatOrEnd(']'); ok {
			panic(catch.Error(`expected end, got "%f"`, f))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = js.ReadFloatOrEnd(']')
	})
	if err != io.ErrUnexpectedEOF {
		t.Fatal("expected ErrUnexpectedEOF")
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		_, _ = js.ReadFloatOrEnd(']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

type testConsumer struct {
	t *testing.T
	m string
	i int64
}

func (t *testConsumer) UnmarshalFromJSON(js Decoder, firstToken json.Token) {
	AssertDelim(firstToken, '{')
	var s string
	ok := true
	for ok {
		if s, ok = js.ReadStringOrEnd('}'); ok {
			switch s {
			case "m":
				t.m = js.ReadString()
			case "i":
				t.i = js.ReadInt()
			default:
				t.t.Fatalf("unexpected string %q", s)
			}
		}
	}
}

func TestReadConsumer(t *testing.T) {
	js := decoderOn(`[{"m":"message","i":42}, null]`)
	err := catch.Do(func() {
		js.ReadDelim('[')
		tc := &testConsumer{t: t}
		valid := js.ReadConsumer(tc)
		if !(valid && tc.m == "message" && tc.i == 42) {
			t.Fatal("unexpected consumer values")
		}
		valid = js.ReadConsumer(tc)
		if valid {
			t.Fatal("expected null, got valid consumer")
		}
		js.ReadDelim(']')
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		tc := &testConsumer{t: t}
		js.ReadConsumer(tc)
	})
	if err != io.ErrUnexpectedEOF {
		t.Fatal("expected ErrUnexpectedEOF")
	}
}

func TestReadConsumerOrEnd(t *testing.T) {
	js := decoderOn(`[{"m":"message","i":42}, null]`)
	err := catch.Do(func() {
		js.ReadDelim('[')
		tc := &testConsumer{t: t}
		valid, ok := js.ReadConsumerOrEnd(tc, ']')
		if !(ok && valid) {
			panic(catch.Error(`expected consumer, got end`))
		}
		valid, ok = js.ReadConsumerOrEnd(tc, ']')
		if ok {
			if valid {
				panic(catch.Error(`expected null, got valid consumer`))
			}
		} else {
			panic(catch.Error(`expected null, got end`))
		}
		valid, ok = js.ReadConsumerOrEnd(tc, ']')
		if ok {
			if valid {
				panic(catch.Error(`expected end, got valid consumer`))
			}
			panic(catch.Error(`expected end, got null`))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		tc := &testConsumer{t: t}
		_, _ = js.ReadConsumerOrEnd(tc, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
