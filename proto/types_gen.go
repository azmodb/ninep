package proto

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tversion) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tversion) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rversion) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rversion) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tauth) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tauth) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rauth) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rauth) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tattach) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tattach) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rattach) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rattach) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rerror) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rerror) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tflush) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tflush) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rflush) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rflush) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Twalk) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Twalk) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rwalk) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rwalk) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Topen) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Topen) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Ropen) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Ropen) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tcreate) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tcreate) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rcreate) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rcreate) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tread) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tread) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rread) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rread) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Twrite) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Twrite) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rwrite) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rwrite) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tclunk) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tclunk) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rclunk) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rclunk) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tremove) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tremove) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rremove) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rremove) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Tstat) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Tstat) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rstat) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rstat) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Twstat) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Twstat) Tag() uint16 { return guint16(m[5:7]) }

// Len returns the total length of the message in bytes. Each 9P message
// begins with a four-byte size field specifying the length in bytes of
// the complete message including the four bytes of the size field
// itself.
func (m Rwstat) Len() int64 { return int64(guint32(m[:4])) }

// Tag is a transaction identifier. No two pending T-messages may use
// the same tag. All R-messages must reference the T-message being
// answered by using the same tag.
func (m Rwstat) Tag() uint16 { return guint16(m[5:7]) }
