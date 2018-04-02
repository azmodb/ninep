package proto

type Tversion []byte

func parseTversion(data []byte) (Tversion, error) {
	if err := verify(data, 11, 0, verifyVersion); err != nil {
		return nil, err
	}
	return Tversion(data), nil
}

func (m Tversion) Msize() uint32   { return guint32(m[7:11]) }
func (m Tversion) Version() string { return string(gfield(m, 11, 0)) }

func (m Tversion) Len() int64  { return int64(guint32(m[:4])) }
func (m Tversion) Tag() uint16 { return guint16(m[5:7]) }

type Rversion []byte

func parseRversion(data []byte) (Rversion, error) {
	if err := verify(data, 11, 0, verifyVersion); err != nil {
		return nil, err
	}
	return Rversion(data), nil
}

func (m Rversion) Msize() uint32   { return guint32(m[7:11]) }
func (m Rversion) Version() string { return string(gfield(m, 11, 0)) }

func (m Rversion) Len() int64  { return int64(guint32(m[:4])) }
func (m Rversion) Tag() uint16 { return guint16(m[5:7]) }

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

func (m Tauth) Afid() uint32  { return guint32(m[7:11]) }
func (m Tauth) Uname() string { return string(gfield(m, 11, 0)) }
func (m Tauth) Aname() string { return string(gfield(m, 11, 1)) }

func (m Tauth) Len() int64  { return int64(guint32(m[:4])) }
func (m Tauth) Tag() uint16 { return guint16(m[5:7]) }

type Rauth []byte

func (m Rauth) Aqid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}

func (m Rauth) Len() int64  { return int64(guint32(m[:4])) }
func (m Rauth) Tag() uint16 { return guint16(m[5:7]) }

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

func (m Tattach) Fid() uint32   { return guint32(m[7:11]) }
func (m Tattach) Afid() uint32  { return guint32(m[11:15]) }
func (m Tattach) Uname() string { return string(gfield(m, 15, 0)) }
func (m Tattach) Aname() string { return string(gfield(m, 15, 1)) }

func (m Tattach) Len() int64  { return int64(guint32(m[:4])) }
func (m Tattach) Tag() uint16 { return guint16(m[5:7]) }

type Rattach []byte

func (m Rattach) Qid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}

func (m Rattach) Len() int64  { return int64(guint32(m[:4])) }
func (m Rattach) Tag() uint16 { return guint16(m[5:7]) }

type Tflush []byte

func (m Tflush) Oldtag() uint16 { return guint16(m[7:9]) }

func (m Tflush) Len() int64  { return int64(guint32(m[:4])) }
func (m Tflush) Tag() uint16 { return guint16(m[5:7]) }

type Rflush []byte

func (m Rflush) Len() int64  { return int64(guint32(m[:4])) }
func (m Rflush) Tag() uint16 { return guint16(m[5:7]) }

type Twalk []byte

func parseTwalk(data []byte) (Twalk, error) {
	n := int(guint16(data[15:17]))
	if n > maxWalkElem {
		return nil, errMaxWalkElem
	}
	for i := 0; i < n; i++ {
		if err := verify(data, 17, i, verifyName); err != nil {
			return nil, err
		}
	}
	return Twalk(data), nil
}

func (m Twalk) Fid() uint32    { return guint32(m[7:11]) }
func (m Twalk) Newfid() uint32 { return guint32(m[11:15]) }

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

func (m Twalk) Len() int64  { return int64(guint32(m[:4])) }
func (m Twalk) Tag() uint16 { return guint16(m[5:7]) }

type Rwalk []byte

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

func (m Rwalk) Len() int64  { return int64(guint32(m[:4])) }
func (m Rwalk) Tag() uint16 { return guint16(m[5:7]) }

type Topen []byte

func (m Topen) Fid() uint32 { return guint32(m[7:11]) }
func (m Topen) Mode() uint8 { return m[11] }

func (m Topen) Len() int64  { return int64(guint32(m[:4])) }
func (m Topen) Tag() uint16 { return guint16(m[5:7]) }

type Ropen []byte

func (m Ropen) Qid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}
func (m Ropen) Iounit() uint32 { return guint32(m[20:24]) }

func (m Ropen) Len() int64  { return int64(guint32(m[:4])) }
func (m Ropen) Tag() uint16 { return guint16(m[5:7]) }

type Tcreate []byte

func parseTcreate(data []byte) (Tcreate, error) {
	if err := verify(data, 11, 0, verifyName); err != nil {
		return nil, err
	}
	return Tcreate(data), nil
}

func (m Tcreate) Fid() uint32  { return guint32(m[7:11]) }
func (m Tcreate) Name() string { return string(gfield(m, 11, 0)) }
func (m Tcreate) Perm() uint32 {
	offset := 11 + 2 + guint16(m[11:13])
	return guint32(m[offset : offset+4])
}
func (m Tcreate) Mode() uint8 {
	return m[len(gfield(m, 11, 0))+17]
}

func (m Tcreate) Len() int64  { return int64(guint32(m[:4])) }
func (m Tcreate) Tag() uint16 { return guint16(m[5:7]) }

type Rcreate []byte

func (m Rcreate) Qid() Qid {
	return Qid{
		Type:    m[7],
		Version: guint32(m[8:12]),
		Path:    guint64(m[12:20]),
	}
}
func (m Rcreate) Iounit() uint32 { return guint32(m[20:24]) }

func (m Rcreate) Len() int64  { return int64(guint32(m[:4])) }
func (m Rcreate) Tag() uint16 { return guint16(m[5:7]) }

type Tread []byte

func (m Tread) Fid() uint32    { return guint32(m[7:11]) }
func (m Tread) Offset() uint64 { return guint64(m[11:19]) }
func (m Tread) Count() uint32  { return guint32(m[19:23]) }

func (m Tread) Len() int64  { return int64(guint32(m[:4])) }
func (m Tread) Tag() uint16 { return guint16(m[5:7]) }

type Rread []byte

func parseRread(data []byte, max int64) (Rread, error) {
	if err := verifyData(data, 7, max); err != nil {
		return nil, err
	}
	return Rread(data), nil
}

func (m Rread) Count() uint32 { return guint32(m[7:11]) }
func (m Rread) Data() []byte  { return m[11:] }

func (m Rread) Len() int64  { return int64(guint32(m[:4])) }
func (m Rread) Tag() uint16 { return guint16(m[5:7]) }

type Twrite []byte

func parseTwrite(data []byte, max int64) (Twrite, error) {
	if err := verifyData(data, 19, max); err != nil {
		return nil, err
	}
	return Twrite(data), nil
}

func (m Twrite) Fid() uint32    { return guint32(m[7:11]) }
func (m Twrite) Offset() uint64 { return guint64(m[11:19]) }
func (m Twrite) Count() uint32  { return guint32(m[19:23]) }
func (m Twrite) Data() []byte   { return m[23:] }

func (m Twrite) Len() int64  { return int64(guint32(m[:4])) }
func (m Twrite) Tag() uint16 { return guint16(m[5:7]) }

type Rwrite []byte

func (m Rwrite) Count() uint32 { return guint32(m[7:11]) }

func (m Rwrite) Len() int64  { return int64(guint32(m[:4])) }
func (m Rwrite) Tag() uint16 { return guint16(m[5:7]) }

type Tclunk []byte

func (m Tclunk) Fid() uint32 { return guint32(m[7:11]) }

func (m Tclunk) Len() int64  { return int64(guint32(m[:4])) }
func (m Tclunk) Tag() uint16 { return guint16(m[5:7]) }

type Rclunk []byte

func (m Rclunk) Len() int64  { return int64(guint32(m[:4])) }
func (m Rclunk) Tag() uint16 { return guint16(m[5:7]) }

type Tremove []byte

func (m Tremove) Fid() uint32 { return guint32(m[7:11]) }

func (m Tremove) Len() int64  { return int64(guint32(m[:4])) }
func (m Tremove) Tag() uint16 { return guint16(m[5:7]) }

type Rremove []byte

func (m Rremove) Len() int64  { return int64(guint32(m[:4])) }
func (m Rremove) Tag() uint16 { return guint16(m[5:7]) }

type Tstat []byte

func (m Tstat) Fid() uint32 { return guint32(m[7:11]) }

func (m Tstat) Len() int64  { return int64(guint32(m[:4])) }
func (m Tstat) Tag() uint16 { return guint16(m[5:7]) }

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

func (m Rstat) Stat() Stat {
	stat := Stat{}
	stat.UnmarshalBinary(m[9:])
	return stat
}

func (m Rstat) Len() int64  { return int64(guint32(m[:4])) }
func (m Rstat) Tag() uint16 { return guint16(m[5:7]) }

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

func (m Twstat) Fid() uint32 { return guint32(m[7:11]) }
func (m Twstat) Stat() Stat {
	stat := Stat{}
	stat.UnmarshalBinary(m[13:])
	return stat
}

func (m Twstat) Len() int64  { return int64(guint32(m[:4])) }
func (m Twstat) Tag() uint16 { return guint16(m[5:7]) }

type Rwstat []byte

func (m Rwstat) Len() int64  { return int64(guint32(m[:4])) }
func (m Rwstat) Tag() uint16 { return guint16(m[5:7]) }

type Rerror []byte

func (m Rerror) Ename() string { return string(gfield(m, 7, 0)) }

func (m Rerror) Len() int64  { return int64(guint32(m[:4])) }
func (m Rerror) Tag() uint16 { return guint16(m[5:7]) }
