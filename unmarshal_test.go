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
	AssertDelimToken(firstToken, '{')
	for {
		s, ok := AssertStringOrEnd(js, '}')
		if !ok {
			break
		}
		if s == "v" {
			t.v = time.Duration(AssertInt(js)) * time.Millisecond
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

func TestAssertDelim(t *testing.T) {
	js := decoderOn("{}")
	err := catch.Do(func() {
		AssertDelim(js, '{')
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		AssertDelim(js, '{')
	})
	if err == nil {
		t.Fatal("expected error")
	}
	err = catch.Do(func() {
		AssertDelim(js, '{')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssertString(t *testing.T) {
	js := decoderOn(`"a"`)
	err := catch.Do(func() {
		if s := AssertString(js); s != "a" {
			panic(catch.Error(`expected "a", got "%s"`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`1`)
	err = catch.Do(func() {
		AssertString(js)
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssertStringOrEnd(t *testing.T) {
	js := decoderOn(`["a"]`)
	err := catch.Do(func() {
		AssertDelim(js, '[')
		if _, ok := AssertStringOrEnd(js, ']'); !ok {
			panic(catch.Error(`expected "a", got end`))
		}
		if s, ok := AssertStringOrEnd(js, ']'); ok {
			panic(catch.Error(`expected end, got "%s"`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = AssertStringOrEnd(js, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`23`)
	err = catch.Do(func() {
		_, _ = AssertStringOrEnd(js, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssertInt(t *testing.T) {
	js := decoderOn(`1`)
	err := catch.Do(func() {
		if s := AssertInt(js); s != 1 {
			panic(catch.Error(`expected 1, got %d`, s))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		AssertInt(js)
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssertIntOrEnd(t *testing.T) {
	js := decoderOn(`[42]`)
	err := catch.Do(func() {
		AssertDelim(js, '[')
		if _, ok := AssertIntOrEnd(js, ']'); !ok {
			panic(catch.Error(`expected "a", got end`))
		}
		if i, ok := AssertIntOrEnd(js, ']'); ok {
			panic(catch.Error(`expected end, got "%d"`, i))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		_, _ = AssertIntOrEnd(js, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
	js = decoderOn(`"a"`)
	err = catch.Do(func() {
		_, _ = AssertIntOrEnd(js, ']')
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
	AssertDelimToken(firstToken, '{')
	var s string
	ok := true
	for ok {
		if s, ok = AssertStringOrEnd(js, '}'); ok {
			switch s {
			case "m":
				t.m = AssertString(js)
			case "i":
				t.i = AssertInt(js)
			default:
				t.t.Fatalf("unexpected string %q", s)
			}
		}
	}
}

func TestAssertConsumer(t *testing.T) {
	js := decoderOn(`{"m":"message","i":42}`)
	err := catch.Do(func() {
		tc := &testConsumer{t: t}
		AssertConsumer(js, tc)
		if !(tc.m == "message" && tc.i == 42) {
			t.Fatal("unexpected consumer values")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		tc := &testConsumer{t: t}
		AssertConsumer(js, tc)
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssertConsumerOrEnd(t *testing.T) {
	js := decoderOn(`[{"m":"message","i":42}]`)
	err := catch.Do(func() {
		AssertDelim(js, '[')
		tc := &testConsumer{t: t}
		if ok := AssertConsumerOrEnd(js, tc, ']'); !ok {
			panic(catch.Error(`expected consumer, got end`))
		}
		if ok := AssertConsumerOrEnd(js, tc, ']'); ok {
			panic(catch.Error(`expected end, got consumer %v`, tc))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = catch.Do(func() {
		tc := &testConsumer{t: t}
		_ = AssertConsumerOrEnd(js, tc, ']')
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
