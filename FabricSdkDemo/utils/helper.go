package utils

import (
	"bytes"
	"encoding/gob"
	"log"
)

func MarshalToBytes(o interface{}) ([]byte, error) {
	var buf bytes.Buffer
	// Create an encoder and send a value
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(o)
	if err != nil {
		log.Fatal("encode:", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnmarshalItem(b []byte, item interface{}) error {
	var buf = bytes.Buffer{}
	buf.Write(b)
	// Create a decoder and receive a value.
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(item)
	if err != nil {
		log.Fatal("decode:", err)
		return err
	}

	return nil
}
