// Package proto provides the definitions and functions used to implement
// the 9P protocol. The parsing routines within make very few assumptions
// or decisions, so that it may be used for a wide variety of
// higher-level packages.
package proto

//go:generate go run internal/generator.go

import (
	"math"

	"github.com/azmodb/ninep/binary"
)

const (
	// NoFid is a reserved fid used in a Tattach request for the afid
	// field, that indicates that the client does not wish to
	// authenticate the session.
	NoFid = math.MaxUint32

	// NoUID indicates a invalid user id.
	NoUID = math.MaxUint32

	// NoGID indicates a invalid group id.
	NoGID = math.MaxUint32

	// NoTag is the tag for Tversion and Rversion messages.
	NoTag = math.MaxUint16
)

// Error represents a 9P protocol error.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrMessageTooLarge = Error("message too large")
	ErrMessageTooSmall = Error("message too small")
)

// Message represents a 9P message and is used to access fields common to
// all 9P messages.
type Message interface {
	// Len returns the length of the message in bytes.
	Len() int

	// Encode encodes to the given binary.Buffer.
	Encode(*binary.Buffer)

	// Decode decodes from the given binary.Buffer.
	Decode(*binary.Buffer)

	// Reset resets all state.
	Reset()
}

// Payloader is a special message which may include an inline payload.
type Payloader interface {
	Message

	// Payload returns the payload for sending.
	Payload() []byte

	// PutPayload sets the decoded payload.
	PutPayload([]byte)

	// FixedLen returns the length of the message in bytes without the
	// count-delimited payload.
	FixedLen() int
}
