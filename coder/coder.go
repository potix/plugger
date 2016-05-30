package coder

import (
	"bytes"
	"encoding/gob"
	"github.com/pkg/errors"
)

type Coder struct {
}

func (c *Coder) Encode(src interface{}, dst *bytes.Buffer) error {
	enc := gob.NewEncoder(dst)
	if err := enc.Encode(src); err != nil {
                return errors.Wrap(err, "encode error")
        }
        return nil
}

func (c *Coder) Decode(src *bytes.Buffer, dst interface{}) error {
	dec := gob.NewDecoder(src)
	if err := dec.Decode(dst); err != nil {
                return errors.Wrap(err, "decode error")
        }
        return nil 
}

func NewCoder() *Coder {
	return new(Coder)
}
