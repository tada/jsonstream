package jsonstream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tada/catch"
)

// A Decoder provides methods to interpret JSON from a stream of tokens provided by a json.Decoder.
type Decoder interface {
	// ReadBool reads next token from the decoder and asserts that it is an boolean or null. The function returns the
	// boolean (or false in case of null) or raises a panic with a catch.Error if an error occurred or if the token
	// didn't match a boolean or null.
	ReadBool() bool

	// ReadBoolOrEnd reads next token from the decoder and asserts that it is either a boolean, null, or a delimiter
	// that matches the given end. The function returns the boolean (or false in case of null) and true if a boolean or
	// null is found or false and false if the delimiter was found. A panic with a catch.Error is raised if neither of
	// those cases are true.
	ReadBoolOrEnd(end byte) (bool, bool)

	// ReadConsumer reads next token from the decoder and, unless that token is null, it passes that token to the given
	// consumers UnmarshalFromJSON and then returns true. If the null token is read, this function returns false
	ReadConsumer(c Consumer) bool

	// ReadConsumerOrEnd reads next token from the decoder and asserts that it is a consumer, null, or a delimiter that
	// matches the given end. The function returns true, true if a consumer is found, false, true if null is found, and
	// false, false if the delimiter was found. A panic with a catch.Error is raised if neither of those cases are true.
	ReadConsumerOrEnd(c Consumer, end byte) (bool, bool)

	// ReadDelim reads next token from the decoder and asserts that it is equal to the given delimiter. A panic
	// with a catch.Error is raised if that is not the case.
	ReadDelim(delim byte)

	// ReadFloatOrEnd reads next token from the decoder and asserts that it is a float, null, or a delimiter that
	// matches the given end. The function returns the float (or 0.0 in case of null) and true if a float was found or 0
	// and false if the delimiter was found. A panic with a catch.Error is raised if neither of those cases are true.
	ReadFloat() float64

	// ReadFloatOrEnd reads next token from the decoder and asserts that it is either an float or a delimiter that
	// matches the given end. The function returns the float and true if an integer is found or 0 and false
	// if the delimiter was found. A panic with a catch.Error is raised if neither of those cases are true.
	ReadFloatOrEnd(end byte) (float64, bool)

	// ReadInt reads next token from the decoder and asserts that it is an integer or null. The function returns the
	// integer (or 0 in case of null) or raises a panic with a catch.Error if an error occurred or if the token didn't
	// match an integer or null.
	ReadInt() int64

	// ReadIntOrEnd reads next token from the decoder and asserts that it is an integer, null, or a delimiter that
	// matches the given end. The function returns the integer (or 0 in case of null) and true if an integer was found
	// or 0 and false if the delimiter was found. A panic with a catch.Error is raised if neither of those cases are
	// true.
	ReadIntOrEnd(end byte) (int64, bool)

	// ReadString reads next token from the decoder and asserts that it is a string or null. The function returns the
	// string (or an empty string in case of null) or raises a panic with a catch.Error if an error occurred or if the
	// token didn't match a string.
	ReadString() string

	// ReadStringOrEnd reads next token from the decoder and asserts that it is a string, null, or a delimiter that
	// matches the given end. The delimiter must be either a '}' or a ']'. The function returns the string (or an empty
	// string in case of null) and true if a string or null is found or an empty string and false if the delimiter was
	// found. A panic with a catch.Error is raised if neither of those cases are true.
	ReadStringOrEnd(end byte) (string, bool)
}

type decoder struct {
	*json.Decoder
}

// A Consumer can initialize itself using a json.Decoder
type Consumer interface {
	// Initialize this instance from a json.Decoder
	UnmarshalFromJSON(js Decoder, firstToken json.Token)
}

// unexpectedError converts io.EOF to io.ErrUnexpectedEOF. All errors are wrapped in a catch.Error before returned
func unexpectedError(err error) error {
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return catch.Error(err)
}

// NewDecoder creates a new Decoder that reads from the given io.Reader.
func NewDecoder(r io.Reader) Decoder {
	js := json.NewDecoder(r)
	js.UseNumber()
	return &decoder{js}
}

// AssertDelim asserts that the given token is equal to the given delimiter. A panic
// with a catch.Error is raised if that is not the case.
func AssertDelim(t json.Token, delim byte) {
	if d, ok := t.(json.Delim); ok {
		s := d.String()
		if len(s) == 1 && s[0] == delim {
			return
		}
	}
	panic(catch.Error("expected delimiter '%c', got %T %v", delim, t, t))
}

// Unmarshal is a helper function that makes it easy for consumers to implement the standard
// json.Unmarshaller interface.
func Unmarshal(c Consumer, bs []byte) error {
	return catch.Do(func() {
		js := NewDecoder(bytes.NewReader(bs))
		js.ReadConsumer(c)
	})
}

// ReadBool reads next token from the decoder and asserts that it is an boolean or null. The function returns the
// boolean (or false in case of null) or raises a panic with a catch.Error if an error occurred or if the token
// didn't match a boolean or null.
func (d *decoder) ReadBool() bool {
	t, err := d.Token()
	if err == nil {
		if t == nil {
			return false
		}
		if v, ok := t.(bool); ok {
			return v
		}
		err = fmt.Errorf("expected an float, got %T %v", t, t)
	}
	panic(unexpectedError(err))
}

// ReadBoolOrEnd reads next token from the decoder and asserts that it is either a boolean, null, or a delimiter
// that matches the given end. The function returns the boolean (or false in case of null) and true if a boolean or
// null is found or false and false if the delimiter was found. A panic with a catch.Error is raised if neither of
// those cases are true.
func (d *decoder) ReadBoolOrEnd(end byte) (bool, bool) {
	t, err := d.Token()
	if err == nil {
		switch t := t.(type) {
		case nil:
			return false, true
		case bool:
			return t, true
		case json.Delim:
			s := t.String()
			if len(s) == 1 && s[0] == end {
				return false, false
			}
		}
		err = fmt.Errorf("expected an boolean or the delimiter '%c' got %T %v", end, t, t)
	}
	panic(unexpectedError(err))
}

// ReadConsumer reads next token from the decoder and, unless that token is null, it passes that token to the given
// consumers UnmarshalFromJSON and then returns true. If the null token is read, this function returns false
func (d *decoder) ReadConsumer(c Consumer) bool {
	t, err := d.Token()
	if err == nil {
		if t == nil {
			return false
		}
		c.UnmarshalFromJSON(d, t)
		return true
	}
	panic(unexpectedError(err))
}

// ReadConsumerOrEnd reads next token from the decoder and asserts that it is a consumer, null, or a delimiter that
// matches the given end. The function returns true, true if a consumer is found, false, true if null is found, and
// false, false if the delimiter was found. A panic with a catch.Error is raised if neither of those cases are true.
func (d *decoder) ReadConsumerOrEnd(c Consumer, end byte) (bool, bool) {
	t, err := d.Token()
	if err == nil {
		if t == nil {
			return false, true
		}
		if d, ok := t.(json.Delim); ok {
			s := d.String()
			if len(s) == 1 && s[0] == end {
				return false, false
			}
		}
		c.UnmarshalFromJSON(d, t)
		return true, true
	}
	panic(unexpectedError(err))
}

// ReadDelim reads next token from the decoder and asserts that it is equal to the given delimiter. A panic
// with a catch.Error is raised if that is not the case.
func (d *decoder) ReadDelim(delim byte) {
	t, err := d.Token()
	if err == nil {
		AssertDelim(t, delim)
		return
	}
	panic(unexpectedError(err))
}

// ReadFloat reads next token from the decoder and asserts that it is a float or null. The function returns the
// float (or 0.0 in case of null) or raises a panic with a catch.Error if an error occurred or if the token didn't
// match a float or null.
func (d *decoder) ReadFloat() float64 {
	t, err := d.Token()
	if err == nil {
		if t == nil {
			return 0
		}
		if s, ok := t.(json.Number); ok {
			var f float64
			if f, err = s.Float64(); err == nil {
				return f
			}
		}
		err = fmt.Errorf("expected an float, got %T %v", t, t)
	}
	panic(unexpectedError(err))
}

// ReadFloatOrEnd reads next token from the decoder and asserts that it is a float, null, or a delimiter that
// matches the given end. The function returns the float (or 0.0 in case of null) and true if a float was found or 0
// and false if the delimiter was found. A panic with a catch.Error is raised if neither of those cases are true.
func (d *decoder) ReadFloatOrEnd(end byte) (float64, bool) {
	t, err := d.Token()
	if err == nil {
		switch t := t.(type) {
		case nil:
			return 0, true
		case json.Number:
			var f float64
			if f, err = t.Float64(); err == nil {
				return f, true
			}
		case json.Delim:
			s := t.String()
			if len(s) == 1 && s[0] == end {
				return 0, false
			}
		}
		err = fmt.Errorf("expected an float or the delimiter '%c' got %T %v", end, t, t)
	}
	panic(unexpectedError(err))
}

// ReadInt reads next token from the decoder and asserts that it is an integer or null. The function returns the
// integer (or 0 in case of null) or raises a panic with a catch.Error if an error occurred or if the token didn't
// match an integer or null.
func (d *decoder) ReadInt() int64 {
	t, err := d.Token()
	if err == nil {
		if t == nil {
			return 0
		}
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

// ReadIntOrEnd reads next token from the decoder and asserts that it is an integer, null, or a delimiter that
// matches the given end. The function returns the integer (or 0 in case of null) and true if an integer was found
// or 0 and false if the delimiter was found. A panic with a catch.Error is raised if neither of those cases are
// true.
func (d *decoder) ReadIntOrEnd(end byte) (int64, bool) {
	t, err := d.Token()
	if err == nil {
		switch t := t.(type) {
		case nil:
			return 0, true
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

// ReadString reads next token from the decoder and asserts that it is a string or null. The function returns the
// string (or an empty string in case of null) or raises a panic with a catch.Error if an error occurred or if the
// token didn't match a string.
func (d *decoder) ReadString() string {
	t, err := d.Token()
	if err == nil {
		if t == nil {
			return ""
		}
		if s, ok := t.(string); ok {
			return s
		}
		err = fmt.Errorf("expected a string, got %T %v", t, t)
	}
	panic(unexpectedError(err))
}

// ReadStringOrEnd reads next token from the decoder and asserts that it is a string, null, or a delimiter that
// matches the given end. The delimiter must be either a '}' or a ']'. The function returns the string (or an empty
// string in case of null) and true if a string or null is found or an empty string and false if the delimiter was
// found. A panic with a catch.Error is raised if neither of those cases are true.
func (d *decoder) ReadStringOrEnd(end byte) (string, bool) {
	t, err := d.Token()
	if err == nil {
		switch t := t.(type) {
		case nil:
			return "", true
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
