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

func decoderOn(s string) *json.Decoder {
	js := json.NewDecoder(bytes.NewReader([]byte(s)))
	js.UseNumber()
	return js
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

func (t *ts) UnmarshalFromJSON(js *json.Decoder, firstToken json.Token) {
	AssertDelim(firstToken, '{')
	for {
		s, ok := ReadStringOrEnd(js, '}')
		if !ok {
			break
		}
		if s == "v" {
			t.v = time.Duration(ReadInt(js)) * time.Millisecond
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
		ReadDelim(js, '{')
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		ReadDelim(js, '{')
	})
	if err == nil {
		t.Fatal("expected error")
	}
	err = catch.Do(func() {
		ReadDelim(js, '{')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadString(t *testing.T) {
	js := decoderOn(`"a"`)
	err := catch.Do(func() {
		if s := ReadString(js); s != "a" {
			panic(catch.Error(`expected "a", got "%s"`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`1`)
	err = catch.Do(func() {
		ReadString(js)
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadStringOrEnd(t *testing.T) {
	js := decoderOn(`["a"]`)
	err := catch.Do(func() {
		ReadDelim(js, '[')
		if _, ok := ReadStringOrEnd(js, ']'); !ok {
			panic(catch.Error(`expected "a", got end`))
		}
		if s, ok := ReadStringOrEnd(js, ']'); ok {
			panic(catch.Error(`expected end, got "%s"`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = ReadStringOrEnd(js, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`23`)
	err = catch.Do(func() {
		_, _ = ReadStringOrEnd(js, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadInt(t *testing.T) {
	js := decoderOn(`1`)
	err := catch.Do(func() {
		if s := ReadInt(js); s != 1 {
			panic(catch.Error(`expected 1, got %d`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		ReadInt(js)
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadIntOrEnd(t *testing.T) {
	js := decoderOn(`[42]`)
	err := catch.Do(func() {
		ReadDelim(js, '[')
		if _, ok := ReadIntOrEnd(js, ']'); !ok {
			panic(catch.Error(`expected "a", got end`))
		}
		if i, ok := ReadIntOrEnd(js, ']'); ok {
			panic(catch.Error(`expected end, got "%d"`, i))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = ReadIntOrEnd(js, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		_, _ = ReadIntOrEnd(js, ']')
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

func (t *testConsumer) UnmarshalFromJSON(js *json.Decoder, firstToken json.Token) {
	AssertDelim(firstToken, '{')
	var s string
	ok := true
	for ok {
		if s, ok = ReadStringOrEnd(js, '}'); ok {
			switch s {
			case "m":
				t.m = ReadString(js)
			case "i":
				t.i = ReadInt(js)
			default:
				t.t.Fatalf("unexpected string %q", s)
			}
		}
	}
}

func TestReadConsumer(t *testing.T) {
	js := decoderOn(`{"m":"message","i":42}`)
	err := catch.Do(func() {
		tc := &testConsumer{t: t}
		ReadConsumer(js, tc)
		if !(tc.m == "message" && tc.i == 42) {
			t.Fatal("unexpected consumer values")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		tc := &testConsumer{t: t}
		ReadConsumer(js, tc)
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadConsumerOrEnd(t *testing.T) {
	js := decoderOn(`[{"m":"message","i":42}]`)
	err := catch.Do(func() {
		ReadDelim(js, '[')
		tc := &testConsumer{t: t}
		if ok := ReadConsumerOrEnd(js, tc, ']'); !ok {
			panic(catch.Error(`expected consumer, got end`))
		}
		if ok := ReadConsumerOrEnd(js, tc, ']'); ok {
			panic(catch.Error(`expected end, got consumer %v`, tc))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		tc := &testConsumer{t: t}
		_ = ReadConsumerOrEnd(js, tc, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
