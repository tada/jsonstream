JSONStream provides helper functions to enable true JSON streaming capabilities.

[![](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![](https://goreportcard.com/badge/github.com/tada/jsonstream)](https://goreportcard.com/report/github.com/tada/jsonstream)
[![](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/tada/jsonstream)
[![](https://github.com/tada/jsonstream/workflows/JSONStream%20Test/badge.svg)](https://github.com/tada/jsonstream/actions)
[![](https://coveralls.io/repos/github/tada/jsonstream/badge.svg?service=github)](https://coveralls.io/github/tada/jsonstream)

### How to get:
```sh
go get github.com/tada/jsonstream
```
### Sample usage

Since jsonstream uses the github.com/tada/catch module, all error handling is baked into the support functions. If any
error is encountered, it will result in a panic that is recovered in the top level functions `Marshal` and `Unmarshal`.
Code using the support functions, like in the example below, is compact since no error propagation is needed.
```go
package tst

import (
  "encoding/json"
  "io"
  "time"

  "github.com/tada/catch/pio"
  "github.com/tada/jsonstream"
)

type ts struct {
  v time.Duration
}

// MarshalJSON is from the json.Marshaller interface
func (t *ts) MarshalJSON() ([]byte, error) {
  // Dispatch to the helper function
  return jsonstream.Marshal(t)
}

// UnmarshalJSON is from the json.Marshaller interface
func (t *ts) UnmarshalJSON(bs []byte) error {
  // Dispatch to the helper function
  return jsonstream.Unmarshal(t, bs)
}

// MarshalToJSON encode as json and stream result to the writer
func (t *ts) MarshalToJSON(w io.Writer) {
  pio.WriteByte('{', w)
  jsonstream.WriteString("v", w)
  pio.WriteByte(':', w)
  pio.WriteInt(int64(t.v/time.Millisecond), w)
  pio.WriteByte('}', w)
}

// UnmarshalToJSON decodes using the given decoder
func (t *ts) UnmarshalFromJSON(js jsonstream.Decoder, firstToken json.Token) {
  jsonstream.AssertDelim(firstToken, '{')
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
```
