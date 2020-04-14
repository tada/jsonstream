package jsonstream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tada/catch"
)

// A Consumer can initialize itself using a json.Decoder
type Consumer interface {
	// Initialize this instance from a json.Decoder
	UnmarshalFromJSON(js *json.Decoder, firstToken json.Token)
}

// unexpectedError converts io.EOF to io.ErrUnexpectedEOF. The argument is returned for all other values
func unexpectedError(err error) error {
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return catch.Error(err)
}

// Unmarshal is a helper function that makes it easy for consumers to implement the standard
// json.Unmarshaller interface.
func Unmarshal(c Consumer, bs []byte) error {
	return catch.Do(func() {
		js := json.NewDecoder(bytes.NewReader(bs))
		js.UseNumber()
		ReadConsumer(js, c)
	})
}

// ReadDelim reads next token from the decoder and asserts that it is equal to the given delimiter. A panic
// with a pio.Error is raised if that is not the case.
func ReadDelim(js *json.Decoder, delim byte) {
	t, err := js.Token()
	if err == nil {
		AssertDelim(t, delim)
		return
	}
	panic(unexpectedError(err))
}

// AssertDelim asserts that the given token is equal to the given delimiter. A panic
// with a pio.Error is raised if that is not the case.
func AssertDelim(t json.Token, delim byte) {
	if d, ok := t.(json.Delim); ok {
		s := d.String()
		if len(s) == 1 && s[0] == delim {
			return
		}
	}
	panic(catch.Error("expected delimiter '%c', got %T %v", delim, t, t))
}

// ReadString reads next token from the decoder and asserts that it is a string. The function returns the string
// or raises a panic with a pio.Error if an error occurred or if the token didn't match a string.
func ReadString(js *json.Decoder) string {
	t, err := js.Token()
	if err == nil {
		if s, ok := t.(string); ok {
			return s
		}
		err = fmt.Errorf("expected a string, got %T %v", t, t)
	}
	panic(unexpectedError(err))
}

// ReadInt reads next token from the decoder and asserts that it is an integer. The function returns the integer
// or raises a panic with a pio.Error if an error occurred or if the token didn't match a string.
//
// The decoder must be configured with UseNumber()
func ReadInt(js *json.Decoder) int64 {
	t, err := js.Token()
	if err == nil {
		if s, ok := t.(json.Number); ok {
			var i int64
			if i, err = s.Int64(); err == nil {
				return i
			}
		}
		err = fmt.Errorf("expected an integer, got %T %v", t, t)
	}
	panic(unexpectedError(err))
}

// ReadStringOrEnd reads next token from the decoder and asserts that it is either a string or a delimiter that
// matches the given end. The delimiter must be either a '}' or a ']'. The function returns the string and true if a
// string is found or an empty string and false if the delimiter was found. A panic with a pio.Error is raised if
// neither of those cases are true.
func ReadStringOrEnd(js *json.Decoder, end byte) (string, bool) {
	t, err := js.Token()
	if err == nil {
		switch t := t.(type) {
		case string:
			return t, true
		case json.Delim:
			s := t.String()
			if len(s) == 1 && s[0] == end {
				return ``, false
			}
		}
		err = fmt.Errorf("expected a string or the delimiter '%c' got %T %v", end, t, t)
	}
	panic(unexpectedError(err))
}

// ReadIntOrEnd reads next token from the decoder and asserts that it is either an integer or a delimiter that
// matches the given end. The function returns the integer and true if an integer is found or 0 and false
// if the delimiter was found. A panic with a pio.Error is raised if neither of those cases are true.
//
// The decoder must be configured with UseNumber()
func ReadIntOrEnd(js *json.Decoder, end byte) (int64, bool) {
	t, err := js.Token()
	if err == nil {
		switch t := t.(type) {
		case json.Number:
			var i int64
			if i, err = t.Int64(); err == nil {
				return i, true
			}
		case json.Delim:
			s := t.String()
			if len(s) == 1 && s[0] == end {
				return 0, false
			}
		}
		err = fmt.Errorf("expected an integer or the delimiter '%c' got %T %v", end, t, t)
	}
	panic(unexpectedError(err))
}

// ReadConsumer reads next token from the decoder and passes that token to the given consumers UnmarshalFromJSON.
func ReadConsumer(js *json.Decoder, c Consumer) {
	t, err := js.Token()
	if err == nil {
		c.UnmarshalFromJSON(js, t)
		return
	}
	panic(unexpectedError(err))
}

// ReadConsumerOrEnd reads next token from the decoder and asserts that it is either an consumer or a delimiter that
// matches the given end. The function returns true if a consumer is found and false if the delimiter was found. A
// panic with a pio.Error is raised if neither of those cases are true.
func ReadConsumerOrEnd(js *json.Decoder, c Consumer, end byte) bool {
	t, err := js.Token()
	if err == nil {
		if d, ok := t.(json.Delim); ok {
			s := d.String()
			if len(s) == 1 && s[0] == end {
				return false
			}
		}
		c.UnmarshalFromJSON(js, t)
		return true
	}
	panic(unexpectedError(err))
}
