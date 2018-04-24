package proto

//go:generate go run generate.go

// Tversion request negotiates the protocol version and message size to be
// used on the connection and initializes the connection for I/O. Tversion
// must be the first message sent on the 9P2000 connection, and the client
// cannot issue any further requests until it has received the Rversion
// reply.
type Tversion []byte

func parseTversion(data []byte) (Tversion, error) {
	if err := verify(data, 11, 0, verifyVersion); err != nil {
		return nil, err
	}
	return Tversion(data), nil
}

// Msize returns the maximum length, in bytes, that the client will ever
// generate or expect to receive in a single 9P2000 message. This count
// includes all 9P2000 protocol data, starting from the size field and
// extending through the message, but excludes enveloping transport
// protocols.
func (m Tversion) Msize() uint32 { return guint32(m[7:11]) }

// Version identifies the level of the protocol that the client supports.
// The string must always begin with the two characters "9P".
func (m Tversion) Version() string { return string(gfield(m, 11, 0)) }

// Rversion reply is sent in response to a Tversion request. It contains
// the version of the protocol that the server has chosen, and the
// maximum size of all successive messages.
type Rversion []byte

func parseRversion(data []byte) (Rversion, error) {
	if err := verify(data, 11, 0, verifyVersion); err != nil {
		return nil, err
	}
	return Rversion(data), nil
}

// Msize returns the maximum size (in bytes) of any 9P200 message that it
// will send or accept, and must be equal to or less than the maximum
// suggested in the preceding Tversion message. After the Rversion message
// is received, both sides of the connection must honor this limit.
func (m Rversion) Msize() uint32 { return guint32(m[7:11]) }

// Version identifies the level of the protocol that the server supports.
// If a server does not understand the protocol version sent in a Tversion
// message, Version will return the string "unknown". A server may choose
// to specify a version that is less than or equal to that supported by
// the client.
func (m Rversion) Version() string { return string(gfield(m, 11, 0)) }

// Tauth message is used to authenticate users on a connection. If the
// server does require authentication, it returns aqid defining a file
// of type QTAUTH that may be read and written to execute an
// authentication protocol. That protocol's definition is not part of
// 9P2000 itself.
//
// Once the protocol is complete, the same afid is presented in the
// attach message for the user, granting entry. The same validated afid
// may be used for multiple attach messages with the same uname and
// aname.
type Tauth []byte

func parseTauth(data []byte) (Tauth, error) {
	if err := verify(data, 11, 0, verifyUname); err != nil {
		return nil, err
	}
	if err := verify(data, 11, 1, verifyPath); err != nil {
		return nil, err
	}
	return Tauth(data), nil
}

// Afid establishes an authentication file.
func (m Tauth) Afid() uint32 { return guint32(m[7:11]) }

// Uname contains the name of the user to authenticate.
func (m Tauth) Uname() string { return string(gfield(m, 11, 0)) }

// Aname contains the name of the file tree to access. It may be empty.
func (m Tauth) Aname() string { return string(gfield(m, 11, 1)) }

// Rauth reply is sent in response to a Tauth request. If a server does
// not require authentication, it can reply to a Tauth message with an
// Rerror message.
type Rauth []byte

// Aqid is returned, if the server does require authentication. The aqid
// of an Rauth message must be of type QTAUTH.
func (m Rauth) Aqid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}

// Tattach message serves as a fresh introduction from a user on the
// client machine to the server.
type Tattach []byte

func parseTattach(data []byte) (Tattach, error) {
	if err := verify(data, 15, 0, verifyUname); err != nil {
		return nil, err
	}
	if err := verify(data, 15, 1, verifyPath); err != nil {
		return nil, err
	}
	return Tattach(data), nil
}

// Fid establishes a fid to be used as the root of the file tree, should
// the client's Tattach request be accepted.
func (m Tattach) Fid() uint32 { return guint32(m[7:11]) }

// Afid serves to authenticate a user, and must have been established in
// a previous Tauth request.i
func (m Tattach) Afid() uint32 { return guint32(m[11:15]) }

// Uname is the user name of the attaching user.
func (m Tattach) Uname() string { return string(gfield(m, 15, 0)) }

// Aname is the name of the file tree that the client wants to access.
// It may be empty.
func (m Tattach) Aname() string { return string(gfield(m, 15, 1)) }

// Rattach message contains a server's reply to a Tattach request. As a
// result of the attach transaction, the client will have a connection to
// the root directory of the desired file tree, represented by the
// returned qid.
type Rattach []byte

// Qid is the qid of the root of the file tree. Qid is associated with
// the fid of the corresponding Tattach request.
func (m Rattach) Qid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}

// Tflush request is sent to the server to purge the pending response.
type Tflush []byte

// Oldtag identifies the message being flushed.
func (m Tflush) Oldtag() uint16 { return guint16(m[7:9]) }

// Rflush echoes the tag (not oldtag) of the Tflush message.
type Rflush []byte

// Twalk message is used to descend a directory hierarchy.
type Twalk []byte

func parseTwalk(data []byte) (Twalk, error) {
	n := int(guint16(data[15:17]))
	if n > MaxWalkElem {
		return nil, errMaxWalkElem
	}
	for i := 0; i < n; i++ {
		if err := verify(data, 17, i, verifyName); err != nil {
			return nil, err
		}
	}
	return Twalk(data), nil
}

// Fid must have been established by a previous transaction, such as an
// Tattach.
func (m Twalk) Fid() uint32 { return guint32(m[7:11]) }

// Newfid contains the proposed fid that the client wishes to associate
// with the result of traversing the directory hierarchy.
func (m Twalk) Newfid() uint32 { return guint32(m[11:15]) }

// Wname contains an ordered list of path name elements that the client
// wishes to descend into in succession.
func (m Twalk) Wname() []string {
	n := int(guint16(m[15:17]))
	if n == 0 {
		return nil
	}

	w := make([]string, 0, n)
	for i := 0; i < n; i++ {
		w = append(w, string(gfield(m, 17, i)))
	}
	return w
}

// Rwalk message contains a server's reply to a successful Twalk request.
// If the first path in the corresponding Twalk request cannot be walked,
// an Rerror message is returned instead.
type Rwalk []byte

// Wqid contains the Qid values of each path in the walk requested by the
// client, up to the first failure.
func (m Rwalk) Wqid() []Qid {
	n := int(guint16(m[7:9]))
	if n == 0 {
		return nil
	}

	qids := make([]Qid, 0, n)
	for i := 0; i < n; i++ {
		q := (m[9+i*13 : 9+(i+1)*13])
		qids = append(qids, Qid{
			Type:    q[0],
			Version: guint32(q[1:5]),
			Path:    guint64(q[5:13]),
		})
	}
	return qids
}

// Topen request asks the file server to check permissions and prepares a
// fid for I/O with subsequent read and write messages.
type Topen []byte

// Fid is the fid of the file to open, as established by a previous
// transaction (such as a succesful Twalk).
func (m Topen) Fid() uint32 { return guint32(m[7:11]) }

// Mode determines the type of I/O, and is checked against the permissions
// for the file:
//
//  0 (OREAD)    read access
//  1 (OWRITE)   write access
//  2 (ORDWR)    read and write access
//  3 (OEXEC)    execute access
//
// If mode has the OTRUNC (0x10) bit set, the file is to be truncated,
// which requires write permission (if the file is append-only, and
// permission is granted, the open succeeds but the file will not be
// truncated)
//
// If the mode has the ORCLOSE (0x40) bit set, the file is to be removed
// when the fid is clunked, which requires permission to remove the file
// from its directory. All other bits in mode should be zero.
//
// It is illegal to write a directory, truncate it, or attempt to remove
// it on close.
func (m Topen) Mode() uint8 { return m[11] }

// Ropen message contains a servers response to a Topen request. An Ropen
// message is only sent if the server determined that the requesting user
// had the proper permissions required for the Topen to succeed, otherwise
// Rerror is returned.
type Ropen []byte

// Qid contains the unique identifier of the opened file.
func (m Ropen) Qid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}

// Iounit field returned by open and create may be zero. If it is not, it
// is the maximum number of bytes that are guaranteed to be read from or
// written to the file without breaking the I/O transfer into multiple
// messages.
func (m Ropen) Iounit() uint32 { return guint32(m[20:24]) }

// Tcreate request asks the file server to create a new file with the
// name supplied, in the directory represented by fid, and requires
// write permission in the directory. The owner of the file is the
// implied user id of the request, the group of the file is the same
// as dir, and the permissions are the value of
//
//     perm & (~0666 | (dir.perm & 0666))
//
// if a regular file is being created and
//
//     perm & (~0777 | (dir.perm & 0777))
//
// if a directory is being created.
type Tcreate []byte

func parseTcreate(data []byte) (Tcreate, error) {
	if err := verify(data, 11, 0, verifyName); err != nil {
		return nil, err
	}
	return Tcreate(data), nil
}

// Fid is the fid of the directory to create the new file in, as
// established by a previous transaction (such as a succesful Twalk).
func (m Tcreate) Fid() uint32 { return guint32(m[7:11]) }

// Name of the newly created file.
func (m Tcreate) Name() string { return string(gfield(m, 11, 0)) }

// Perm represents the permissions of the newly created file. Directories
// are created by setting the DMDIR bit (0x80000000) in the perm.
func (m Tcreate) Perm() uint32 {
	offset := 11 + 2 + guint16(m[11:13])
	return guint32(m[offset : offset+4])
}

// Mode determines the type of I/O. The newly created file is opened
// according to mode, and fid will represent the newly opened file. Mode
// is not checked against the permissions in perm.
func (m Tcreate) Mode() uint8 {
	return m[len(gfield(m, 11, 0))+17]
}

// Rcreate message contains a servers response to a Tcreate request. An
// Rcreate message is only sent if the server determined that the
// requesting user had the proper permissions required for the Tcreate to
// succeed, otherwise Rerror is returned.
type Rcreate []byte

// Qid contains the unique identifier of the created file.
func (m Rcreate) Qid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}

// Iounit field returned by open and create may be zero. If it is not, it
// is the maximum number of bytes that are guaranteed to be read from or
// written to the file without breaking the I/O transfer into multiple
// messages.
func (m Rcreate) Iounit() uint32 { return guint32(m[20:24]) }

// Tread request asks for count bytes of data from the file, which must
// be opened for reading, starting offset bytes after the beginning of
// the file.
type Tread []byte

// Fid is the handle of the file to read from.
func (m Tread) Fid() uint32 { return guint32(m[7:11]) }

// Offset is the starting point in the file from which to begin returning
// data.
func (m Tread) Offset() uint64 { return guint64(m[11:19]) }

// Count is the number of bytes to read from the file.
func (m Tread) Count() uint32 { return guint32(m[19:23]) }

// Rread message returns the bytes requested by a Tread message.
type Rread []byte

func parseRread(data []byte, max int64) (Rread, error) {
	if err := verifyData(data, 7, max); err != nil {
		return nil, err
	}
	return Rread(data), nil
}

// Count is the number of bytes read from the file.
func (m Rread) Count() uint32 { return guint32(m[7:11]) }

// Data returns the bytes requested by a Tread message.
func (m Rread) Data() []byte { return m[11:] }

// Twrite message is sent by a client to write data to a file.
type Twrite []byte

func parseTwrite(data []byte, max int64) (Twrite, error) {
	if err := verifyData(data, 19, max); err != nil {
		return nil, err
	}
	return Twrite(data), nil
}

// Fid is the handle of the file to write to.
func (m Twrite) Fid() uint32 { return guint32(m[7:11]) }

// Offset is the starting point in the file from which to begin writing
// data.
func (m Twrite) Offset() uint64 { return guint64(m[11:19]) }

// Count is the number of bytes write to the file.
func (m Twrite) Count() uint32 { return guint32(m[19:23]) }

// Data to write to a file.
func (m Twrite) Data() []byte { return m[23:] }

// Rwrite message returns the bytes requested by a Twrite message.
type Rwrite []byte

// Count is the number of bytes written to a file.
func (m Rwrite) Count() uint32 { return guint32(m[7:11]) }

// Tclunk request informs the file server that the current file
// represented by fid is no longer needed by the client. The actual file
// is not removed on the server unless the fid had been opened with
// ORCLOSE.
type Tclunk []byte

// Fid represents the no longer needed file.
func (m Tclunk) Fid() uint32 { return guint32(m[7:11]) }

// Rclunk message contains a servers response to a Tclunk request.
type Rclunk []byte

// Tremove request asks the file server both to remove a file and to
// clunk it, even if the remove fails. This request will fail if the
// client does not have write permission in the parent directory.
type Tremove []byte

// Fid is the handle of the file to remove.
func (m Tremove) Fid() uint32 { return guint32(m[7:11]) }

// Rremove message contains a servers response to a Tremove request. An
// Rremove message is only sent if the server determined that the
// requesting user had the proper permissions required for the Tremove to
// succeed, otherwise Rerror is returned.
type Rremove []byte

// Tstat message inquires about a file.
type Tstat []byte

// Fid is the handle of the file to inquire about.
func (m Tstat) Fid() uint32 { return guint32(m[7:11]) }

// Rstat message contains a servers response to a Tstat request.
type Rstat []byte

func parseRstat(data []byte) (Rstat, error) {
	n := headerLen + 2 + 2 + fixedStatLen - (2 * 4)
	if err := verify(data, n, 0, verifyName); err != nil {
		return nil, err
	}
	if err := verify(data, n, 1, verifyUname); err != nil {
		return nil, err
	}
	if err := verify(data, n, 2, verifyUname); err != nil {
		return nil, err
	}
	if err := verify(data, n, 3, verifyUname); err != nil {
		return nil, err
	}
	return Rstat(data), nil
}

// Stat returns the machine-independent directory entry.
func (m Rstat) Stat() Stat { return gstat(m[9:]) }

// Twstat request can change some of the file status information. The
// name can be changed by anyone with write permission in the parent
// directory. It is an error to change the name to that of an existing
// file.
// The length can be changed (affecting the actual length of the file)
// by anyone with write permission on the file. It is an error to
// attempt to set the length of a directory to a non-zero value, and
// servers may decide to reject length changes for other reasons. The
// mode and mtime can be changed by the owner of the file or the group
// leader of the file's current group. The directory bit cannot be
// changed by a Twstat. The other defined permission and mode bits can.
// The gid can be changed: by the owner if also a member of the new
// group; or by the group leader of the file's current group if also
// leader of the new group.  None of the other data can be altered by a
// Twstat request and attempts to change them will trigger an error.
//
// Either all the changes in wstat request happen, or none of them does.
//
// A Twstat request can avoid modifying some properties of the file by
// providing explicit `don't touch` values in the stat data that is sent:
// zero-length strings for text values and the maximum unsigned value of
// appropriate size for integral values. As a special case, if all the
// elements of the directory entry in a Twstat message are `don't touch`
// values, the server may interpret it as a request to guarantee that
// the contents of the associated file are committed to stable storage
// before the Rwstat message is returned.
type Twstat []byte

func parseTwstat(data []byte) (Twstat, error) {
	n := headerLen + 4 + 2 + 2 + fixedStatLen - (2 * 4)
	if err := verify(data, n, 0, verifyName); err != nil {
		return nil, err
	}
	if err := verify(data, n, 1, verifyUname); err != nil {
		return nil, err
	}
	if err := verify(data, n, 2, verifyUname); err != nil {
		return nil, err
	}
	if err := verify(data, n, 3, verifyUname); err != nil {
		return nil, err
	}
	return Twstat(data), nil
}

// Fid is the handle of the file to change.
func (m Twstat) Fid() uint32 { return guint32(m[7:11]) }

// Stat returns the machine-independent directory entry.
func (m Twstat) Stat() Stat { return gstat(m[13:]) }

// Rwstat message contains a servers response to a Twstat request. An
// Rwstat message is only sent if the server determined that the
// requesting user had the proper permissions required for the Twstat to
// succeed, otherwise Rerror is returned.
type Rwstat []byte

// Rerror message (there is no Terror) is used to return an error string
// describing the failure of a transaction.
type Rerror []byte

// Ename is a UTF-8 string describing the error that occured.
func (m Rerror) Ename() string { return string(gfield(m, 7, 0)) }
