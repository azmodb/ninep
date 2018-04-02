// +build gofuzz

package proto

import "bytes"

func Fuzz(data []byte) int {
	dec := NewDecoder(bytes.NewReader(data))
	if _, err := dec.Decode(); err != nil {
		return 0
	}
	return 1
}
