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
	// FixedReadWriteSize is the length of all fixed-width fields in a
	// Twrite or Tread message. Twrite and Tread messages are defined as
	//
	//     size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	//     size[4] Tread  tag[2] fid[4] offset[8] count[4]
	//
	FixedReadWriteSize = HeaderSize + 4 + 8 + 4 // 23

	// MaxDataSize is the maximum data size of a Twrite or Rread message.
	MaxDataSize = math.MaxInt32 - (FixedReadWriteSize + 1) // ~ 2GB

	// MaxNameSize is the maximum length of a filename/username in bytes.
	MaxNameSize = math.MaxUint8

	// MaxNames is the maximum allowed number of path elements in a Twalk
	// request.
	MaxNames = 16

	// MaxMessageSize is the maximum size of a 9P message (Twrite).
	MaxMessageSize = (FixedReadWriteSize + 1) + MaxDataSize

	// MinMessageSize is the minimum size of a 9P Message (Twalk).
	MinMessageSize = HeaderSize + 4 + 4 + 2 + MaxNames*(2+MaxNameSize)

	// DefaultMaxDataSize is the default maximum data size of a Twrite or
	// Rread message.
	DefaultMaxDataSize = 2 * 1024 * 1024 // ~ 2MB

	// DefaultMaxMessageSize is the default maximum size of a 9P2000.L
	// message.
	DefaultMaxMessageSize = (FixedReadWriteSize + 1) + DefaultMaxDataSize

	// HeaderSize is the number of bytes required for a header.
	HeaderSize = 7
)

const (
	// NoFid is a reserved fid used in a Tattach request for the afid
	// field, that indicates that the client does not wish to
	// authenticate the session.
	NoFid = math.MaxUint32

	// NoUid indicates a invalid user id.
	NoUid = math.MaxUint32

	// NoGid indicates a invalid group id.
	NoGid = math.MaxUint32

	// NoTag is the tag for Tversion and Rversion messages.
	NoTag = math.MaxUint16
)

// Error represents a 9P protocol error.
type Error string

func (e Error) Error() string { return string(e) }

const (
	// ErrMessageTooLarge is returned during the parsing process if a
	// message exceeds the maximum size negotiated during the
	// Tversion/Rversion transaction.
	ErrMessageTooLarge = Error("message too large")

	// ErrMessageTooSmall is returned during the parsing process if a
	// message is too small.
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
