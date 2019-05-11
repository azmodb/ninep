package ninep

import (
	"context"
	"testing"
)

const compatTestServerAddr = "127.0.0.1:5640"

func TestCompatVersion(t *testing.T) {
	c, err := Dial(context.Background(), "tcp", compatTestServerAddr)
	if err != nil {
		t.Fatalf("cannot dial compat server: %v", err)
	}
	c.Close()
}
