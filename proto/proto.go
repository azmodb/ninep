package proto

import (
	"unicode/utf8"

	"bytes"
)

// Based on http://9p.io/sources/plan9/sys/include/fcall.h
const (
	msgTversion = iota + 100 // size[4] Tversion tag[2] msize[4] version[s]
	msgRversion              // size[4] Rversion tag[2] msize[4] version[s]
	msgTauth                 // size[4] Tauth tag[2] afid[4] uname[s] aname[s]
	msgRauth                 // size[4] Rauth tag[2] aqid[13]
	msgTattach               // size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
	msgRattach               // size[4] Rattach tag[2] qid[13]
	msgTerror                // illegal
	msgRerror                // size[4] Rerror tag[2] ename[s]
	msgTflush                // size[4] Tflush tag[2] oldtag[2]
	msgRflush                // size[4] Rflush tag[2]
	msgTwalk                 // size[4] Twalk tag[2] fid[4] newfid[4] nwname[2] nwname*(wname[s])
	msgRwalk                 // size[4] Rwalk tag[2] nwqid[2] nwqid*(wqid[13])
	msgTopen                 // size[4] Topen tag[2] fid[4] mode[1]
	msgRopen                 // size[4] Ropen tag[2] qid[13] iounit[4]
	msgTcreate               // size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
	msgRcreate               // size[4] Rcreate tag[2] qid[13] iounit[4]
	msgTread                 // size[4] Tread tag[2] fid[4] offset[8] count[4]
	msgRread                 // size[4] Rread tag[2] count[4] data[count]
	msgTwrite                // size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	msgRwrite                // size[4] Rwrite tag[2] count[4]
	msgTclunk                // size[4] Tclunk tag[2] fid[4]
	msgRclunk                // size[4] Rclunk tag[2]
	msgTremove               // size[4] Tremove tag[2] fid[4]
	msgRremove               // size[4] Rremove tag[2]
	msgTstat                 // size[4] Tstat tag[2] fid[4]
	msgRstat                 // size[4] Rstat tag[2] stat[n]
	msgTwstat                // size[4] Twstat tag[2] fid[4] stat[n]
	msgRwstat                // size[4] Rwstat tag[2]
)

var minSizeLUT = [28]uint32{
	13,                // size[4] Tversion tag[2] msize[4] version[s]
	13,                // size[4] Rversion tag[2] mversion[s]
	15,                // size[4] Tauth tag[2] afid[4] uname[s] aname[s]
	20,                // size[4] Rauth tag[2] aqid[13]
	19,                // size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
	20,                // size[4] Rattach tag[2] qid[13]
	0,                 // illegal
	9,                 // size[4] Rerror tag[2] ename[s]
	9,                 // size[4] Tflush tag[2] oldtag[2]
	7,                 // size[4] Rflush tag[2]
	17,                // size[4] Twalk tag[2] fid[4] newfid[4] nwname[2] nwname*(wname[s])
	9,                 // size[4] Rwalk tag[2] nwqid[2] nwqid*(wqid[13])
	12,                // size[4] Topen tag[2] fid[4] mode[1]
	24,                // size[4] Ropen tag[2] qid[13] iounit[4]
	18,                // size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
	24,                // size[4] Rcreate tag[2] qid[13] iounit[4]
	fixedReadWriteLen, // size[4] Tread tag[2] fid[4] offset[8] count[4]
	11,                // size[4] Rread tag[2] count[4] data[count]
	fixedReadWriteLen, // size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	11,                // size[4] Rwrite tag[2] count[4]
	11,                // size[4] Tclunk tag[2] fid[4]
	7,                 // size[4] Rclunk tag[2]
	11,                // size[4] Tremove tag[2] fid[4]
	7,                 // size[4] Rremove tag[2]
	11,                // size[4] Tstat tag[2] fid[4]
	11 + fixedStatLen, // size[4] Rstat tag[2] stat[n]
	15 + fixedStatLen, // size[4] Twstat tag[2] fid[4] stat[n]
	7,                 // size[4] Rwstat tag[2]
}

var minSizeLUT64 [28]int64

func init() {
	for i, s := range minSizeLUT {
		minSizeLUT64[i] = int64(s)
	}
}

const (
	errElemInvalidUTF8 = Error("path element name is not valid utf8")
	errMaxNames        = Error("maximum walk elements exceeded")
	errElemTooLarge    = Error("path element name too large")
	errPathTooLarge    = Error("file tree name too large")
	errPathName        = Error("separator in path element")

	errUnameInvalidUTF8   = Error("username is not valid utf8")
	errUnameTooLarge      = Error("username is too large")
	errVersionInvalidUTF8 = Error("version is not valid utf8")
	errVersionTooLarge    = Error("version is too large")

	errDataTooLarge = Error("maximum data bytes exeeded")

	errMaxName = Error("maximum walk elements exceeded")

	errInvalidMessageType = Error("invalid message type")
	errMessageTooLarge    = Error("message too large")
	errMessageTooSmall    = Error("message too small")
)

var separator = []byte("/")

const separatorByte = '/'

func verifyUname(uname []byte) error {
	if len(uname) > maxUnameLen {
		return errUnameTooLarge
	}
	if !utf8.Valid(uname) {
		return errUnameInvalidUTF8
	}
	return nil
}

func verifyVersion(version []byte) error {
	if len(version) > maxVersionLen {
		return errVersionTooLarge
	}
	if !utf8.Valid(version) {
		return errVersionInvalidUTF8
	}
	return nil
}

func verifyName(name []byte) error {
	if len(name) > maxNameLen {
		return errElemTooLarge
	}
	if !utf8.Valid(name) {
		return errElemInvalidUTF8
	}
	if bytes.Contains(name, separator) {
		return errPathName
	}
	return nil
}

func verifyPath(name []byte) (err error) {
	for len(name) > 0 && name[0] == separatorByte {
		name = name[1:]
	}
	if len(name) == 0 {
		return nil
	}
	if len(name) > maxPathLen {
		return errPathTooLarge
	}
	if bytes.Count(name, separator) > maxName {
		return errMaxNames
	}

	elems := bytes.Split(name, separator)
	for _, elem := range elems {
		if err = verifyName(elem); err != nil {
			return err
		}
	}
	return nil
}
