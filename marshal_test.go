package jsonstream

import (
	"bytes"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	tv := ts{v: time.Millisecond * 23}
	bs, err := Marshal(&tv)
	if err != nil {
		t.Fatal(err)
	}
	ex := `{"v":23}`
	if !bytes.Equal(bs, []byte(ex)) {
		t.Fatalf("Marshal(): expected: %s, got %s", ex, string(bs))
	}
}

func TestWriteString(t *testing.T) {
	b := bytes.Buffer{}
	WriteString(`The "quoted" part`, &b)
	a := b.String()
	e := `"The \"quoted\" part"`
	if a != e {
		t.Fatalf("WriteString(): expected: %s, got %s", e, a)
	}
}
