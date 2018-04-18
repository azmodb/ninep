// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
)

var methodDoc = []string{
	`// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.`,

	`// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.`,
}

var messages = []string{
	"Tversion",
	"Rversion",
	"Tauth",
	"Rauth",
	"Tattach",
	"Rattach",
	"Terror",
	"Rerror",
	"Tflush",
	"Rflush",
	"Twalk",
	"Rwalk",
	"Topen",
	"Ropen",
	"Tcreate",
	"Rcreate",
	"Tread",
	"Rread",
	"Twrite",
	"Rwrite",
	"Tclunk",
	"Rclunk",
	"Tremove",
	"Rremove",
	"Tstat",
	"Rstat",
	"Twstat",
	"Rwstat",
}

func genmethods(w io.Writer) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "package proto\n\n")

	for _, m := range messages {
		if m == "Terror" {
			continue
		}
		fmt.Fprintf(buf, "%s\n", methodDoc[0])
		fmt.Fprintf(buf, "func (m %s) Len() int64  { return int64(guint32(m[:4])) }\n\n", m)
		fmt.Fprintf(buf, "%s\n", methodDoc[1])
		fmt.Fprintf(buf, "func (m %s) Tag() uint16 { return guint16(m[5:7]) }\n\n", m)
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("genmethods: format source: %v", err)
	}
	if _, err = w.Write(data); err != nil {
		log.Fatalf("genmethods: write output file: %v", err)
	}
}

func main() {
	f, err := os.Create("types_gen.go")
	if err != nil {
		log.Fatalf("cannot create output file: %v", err)
	}
	defer f.Close()

	genmethods(f)
}
