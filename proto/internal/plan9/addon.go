package plan9

func NewVersion(typ uint8, tag uint16, msize uint32, version string) Fcall {
	return Fcall{Type: typ, Tag: tag, Msize: msize, Version: version}
}

func NewTauth(tag uint16, afid uint32, uname, aname string) Fcall {
	return Fcall{Type: Tauth, Tag: tag, Afid: afid, Uname: uname, Aname: aname}
}

func NewRauth(tag uint16, aqid Qid) Fcall {
	return Fcall{Type: Rauth, Aqid: aqid}
}

func NewTattach(tag uint16, fid, afid uint32, uname, aname string) Fcall {
	return Fcall{Type: Tattach, Tag: tag, Fid: fid, Afid: afid, Uname: uname, Aname: aname}
}

func NewRattach(tag uint16, qid Qid) Fcall {
	return Fcall{Type: Rattach, Qid: qid}
}

func NewRerror(tag uint16, ename string) Fcall {
	return Fcall{Type: Rerror, Ename: ename}
}

func NewTflush(tag uint16, oldtag uint16) Fcall {
	return Fcall{Type: Tflush, Tag: tag, Oldtag: oldtag}
}

func NewRflush(tag uint16) Fcall { return Fcall{Type: Rflush, Tag: tag} }

func NewTwalk(tag uint16, fid, newfid uint32, wname []string) Fcall {
	return Fcall{Type: Twalk, Fid: fid, Newfid: newfid, Wname: wname}
}

func NewRwalk(tag uint16, wqid []Qid) Fcall {
	return Fcall{Type: Rwalk, Wqid: wqid}
}

func NewTopen(tag uint16, fid uint32, mode uint8) Fcall {
	return Fcall{Type: Topen, Fid: fid, Mode: mode}
}

func NewRopen(tag uint16, qid Qid, iounit uint32) Fcall {
	return Fcall{Type: Ropen, Qid: qid, Iounit: iounit}
}

func NewTcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) Fcall {
	return Fcall{Type: Tcreate, Fid: fid, Name: name, Perm: Perm(perm), Mode: mode}
}

func NewRcreate(tag uint16, qid Qid, iounit uint32) Fcall {
	return Fcall{Type: Rcreate, Qid: qid, Iounit: iounit}
}

func NewTread(tag uint16, fid uint32, offset uint64, count uint32) Fcall {
	return Fcall{Type: Tread, Tag: tag, Fid: fid, Offset: offset, Count: count}
}

func NewRread(tag uint16, data []byte) Fcall {
	return Fcall{Type: Rread, Tag: tag, Count: uint32(len(data)), Data: data}
}

func NewTwrite(tag uint16, fid uint32, offset uint64, data []byte) Fcall {
	return Fcall{Type: Twrite, Tag: tag, Fid: fid, Offset: offset, Count: uint32(len(data)), Data: data}
}

func NewRwrite(tag uint16, count uint32) Fcall {
	return Fcall{Type: Rwrite, Tag: tag, Count: count}
}

func NewTclunk(tag uint16, fid uint32) Fcall {
	return Fcall{Type: Tclunk, Tag: tag, Fid: fid}
}

func NewRclunk(tag uint16) Fcall { return Fcall{Type: Rclunk, Tag: tag} }

func NewTremove(tag uint16, fid uint32) Fcall {
	return Fcall{Type: Tremove, Tag: tag, Fid: fid}
}

func NewRremove(tag uint16) Fcall { return Fcall{Type: Rremove, Tag: tag} }

func NewTstat(tag uint16, fid uint32) Fcall {
	return Fcall{Type: Tstat, Tag: tag, Fid: fid}
}

func NewRstat(tag uint16, stat []byte) Fcall {
	return Fcall{Type: Rstat, Tag: tag, Stat: stat}
}

func NewTwstat(tag uint16, fid uint32, stat []byte) Fcall {
	return Fcall{Type: Twstat, Tag: tag, Fid: fid, Stat: stat}
}

func NewRwstat(tag uint16) Fcall { return Fcall{Type: Rwstat, Tag: tag} }
