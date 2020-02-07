package main

import (
	"bytes"
	"encoding/gob"
)

func deepCopy(msg M) (M, error) {
	gob.Register(M{})
	gob.Register([]M{})
	gob.Register([]interface{}{})

	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	d := gob.NewDecoder(&b)
	if err := e.Encode(msg); err != nil {
		return nil, err
	}
	r := map[string]interface{}{}
	if err := d.Decode(&r); err != nil {
		return nil, err
	}
	return r, nil
}