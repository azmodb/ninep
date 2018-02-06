package proto

import "fmt"

type Tversion []byte

func parseTversion(data []byte) (Tversion, error) {
	if err := vfield(data, 11, 0, verifyVersion); err != nil {
		return nil, err
	}
	return Tversion(data), nil
}

func (m Tversion) Msize() uint32   { return guint32(m[7:11]) }
func (m Tversion) Version() string { return string(gfield(m, 11, 0)) }

func (m Tversion) Len() int64  { return int64(guint32(m[:4])) }
func (m Tversion) Tag() uint16 { return guint16(m[5:7]) }
func (m Tversion) Reset()      {}
func (m Tversion) String() string {
	return fmt.Sprintf("Tversion tag:%d msize:%d version:%q",
		m.Tag(), m.Msize(), gfield(m, 11, 0))
}

type Rversion []byte

func parseRversion(data []byte) (Rversion, error) {
	if err := vfield(data, 11, 0, verifyVersion); err != nil {
		return nil, err
	}
	return Rversion(data), nil
}

func (m Rversion) Msize() uint32   { return guint32(m[7:11]) }
func (m Rversion) Version() string { return string(gfield(m, 11, 0)) }

func (m Rversion) Len() int64  { return int64(guint32(m[:4])) }
func (m Rversion) Tag() uint16 { return guint16(m[5:7]) }
func (m Rversion) Reset()      {}
func (m Rversion) String() string {
	return fmt.Sprintf("Rversion tag:%d msize:%d version:%q",
		m.Tag(), m.Msize(), gfield(m, 11, 0))
}

type Tauth []byte

func parseTauth(data []byte) (Tauth, error) {
	if err := vfield(data, 11, 0, verifyUname); err != nil {
		return nil, err
	}
	if err := vfield(data, 11, 1, verifyPath); err != nil {
		return nil, err
	}
	return Tauth(data), nil
}

func (m Tauth) Afid() uint32  { return guint32(m[7:11]) }
func (m Tauth) Uname() string { return string(gfield(m, 11, 0)) }
func (m Tauth) Aname() string { return string(gfield(m, 11, 1)) }

func (m Tauth) Len() int64  { return int64(guint32(m[:4])) }
func (m Tauth) Tag() uint16 { return guint16(m[5:7]) }
func (m Tauth) Reset()      {}
func (m Tauth) String() string {
	return fmt.Sprintf("Tauth tag:%d afid:%d uname:%q aname:%q",
		m.Tag(), m.Afid(), gfield(m, 11, 0), gfield(m, 11, 1))
}

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
func (m Rauth) Reset()      {}
func (m Rauth) String() string {
	return fmt.Sprintf("Rauth tag:%d aqid:%q", m.Tag(), m.Aqid())
}

type Tattach []byte

func parseTattach(data []byte) (Tattach, error) {
	if err := vfield(data, 15, 0, verifyUname); err != nil {
		return nil, err
	}
	if err := vfield(data, 15, 1, verifyPath); err != nil {
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
func (m Tattach) Reset()      {}
func (m Tattach) String() string {
	return fmt.Sprintf("Tattach tag:%d fid:%d afid:%d uname:%q aname:%q",
		m.Tag(), m.Fid(), m.Afid(), gfield(m, 15, 0), gfield(m, 15, 1))
}

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
func (m Rattach) Reset()      {}
func (m Rattach) String() string {
	return fmt.Sprintf("Rattach tag:%d qid:%q", m.Tag(), m.Qid())
}

type Tflush []byte

func (m Tflush) Oldtag() uint16 { return guint16(m[7:9]) }

func (m Tflush) Len() int64  { return int64(guint32(m[:4])) }
func (m Tflush) Tag() uint16 { return guint16(m[5:7]) }
func (m Tflush) Reset()      {}
func (m Tflush) String() string {
	return fmt.Sprintf("Tflush tag:%d oldtag:%d",
		m.Tag(), m.Oldtag())
}

type Rflush []byte

func (m Rflush) Len() int64  { return int64(guint32(m[:4])) }
func (m Rflush) Tag() uint16 { return guint16(m[5:7]) }
func (m Rflush) Reset()      {}
func (m Rflush) String() string {
	return fmt.Sprintf("Rflush tag:%d", m.Tag())
}

type Twalk []byte

func parseTwalk(data []byte) (Twalk, error) {
	n := int(guint16(data[15:17]))
	if n > maxName {
		return nil, errMaxName
	}
	for i := 0; i < n; i++ {
		if err := vfield(data, 17, i, verifyName); err != nil {
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
func (m Twalk) Reset()      {}
func (m Twalk) String() string {
	n := int(guint16(m[7:9]))
	w := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		w = append(w, gfield(m, 17, i))
	}

	return fmt.Sprintf("Twalk tag:%d fid:%d newfid:%d wname:%q",
		m.Tag(), m.Fid(), m.Newfid(), w)
}

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
func (m Rwalk) Reset()      {}
func (m Rwalk) String() string {
	return fmt.Sprintf("Rwalk tag:%d wqid:%q", m.Tag(), m.Wqid())
}

type Topen []byte

func (m Topen) Fid() uint32 { return guint32(m[7:11]) }
func (m Topen) Mode() uint8 { return m[11] }

func (m Topen) Len() int64  { return int64(guint32(m[:4])) }
func (m Topen) Tag() uint16 { return guint16(m[5:7]) }
func (m Topen) Reset()      {}
func (m Topen) String() string {
	return fmt.Sprintf("Topen tag:%d fid:%d mode:%d",
		m.Tag(), m.Fid(), m.Mode())
}

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
func (m Ropen) Reset()      {}
func (m Ropen) String() string {
	return fmt.Sprintf("Ropen tag:%d qid:%q iounit:%d",
		m.Tag(), m.Qid(), m.Iounit())
}

type Tcreate []byte

func parseTcreate(data []byte) (Tcreate, error) {
	if err := vfield(data, 11, 0, verifyName); err != nil {
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
func (m Tcreate) Reset()      {}
func (m Tcreate) String() string {
	return fmt.Sprintf("Tcreate tag:%d fid:%d name:%q perm:%d mode:%d",
		m.Tag(), m.Fid(), m.Name(), m.Perm(), m.Mode())
}

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
func (m Rcreate) Reset()      {}
func (m Rcreate) String() string {
	return fmt.Sprintf("Rcreate tag:%d qid:%q iounit:%d",
		m.Tag(), m.Qid(), m.Iounit())
}

type Tread []byte

func (m Tread) Fid() uint32    { return guint32(m[7:11]) }
func (m Tread) Offset() uint64 { return guint64(m[11:19]) }
func (m Tread) Count() uint32  { return guint32(m[19:23]) }

func (m Tread) Len() int64  { return int64(guint32(m[:4])) }
func (m Tread) Tag() uint16 { return guint16(m[5:7]) }
func (m Tread) Reset()      {}
func (m Tread) String() string {
	return fmt.Sprintf("Tread tag:%d fid:%d offset:%d count:%d",
		m.Tag(), m.Fid(), m.Offset(), m.Count())
}

type Rread []byte

func parseRread(data []byte, max int64) (Rread, error) {
	if err := vdata(data, 7, max); err != nil {
		return nil, err
	}
	return Rread(data), nil
}

func (m Rread) Count() uint32 { return guint32(m[7:11]) }
func (m Rread) Data() []byte  { return m[11:] }

func (m Rread) Len() int64  { return int64(guint32(m[:4])) }
func (m Rread) Tag() uint16 { return guint16(m[5:7]) }
func (m Rread) Reset()      {}
func (m Rread) String() string {
	return fmt.Sprintf("Rread tag:%d count:%d", m.Tag(), m.Count())
}

type Twrite []byte

func parseTwrite(data []byte, max int64) (Twrite, error) {
	if err := vdata(data, 19, max); err != nil {
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
func (m Twrite) Reset()      {}
func (m Twrite) String() string {
	return fmt.Sprintf("Twrite tag:%d fid:%d offset:%d count:%d",
		m.Tag(), m.Fid(), m.Offset(), m.Count())
}

type Rwrite []byte

func (m Rwrite) Count() uint32 { return guint32(m[7:11]) }

func (m Rwrite) Len() int64  { return int64(guint32(m[:4])) }
func (m Rwrite) Tag() uint16 { return guint16(m[5:7]) }
func (m Rwrite) Reset()      {}
func (m Rwrite) String() string {
	return fmt.Sprintf("Rwrite tag:%d count:%d", m.Tag(), m.Count())
}

type Tclunk []byte

func (m Tclunk) Fid() uint32 { return guint32(m[7:11]) }

func (m Tclunk) Len() int64  { return int64(guint32(m[:4])) }
func (m Tclunk) Tag() uint16 { return guint16(m[5:7]) }
func (m Tclunk) Reset()      {}
func (m Tclunk) String() string {
	return fmt.Sprintf("Tclunk tag:%d fid:%d", m.Tag(), m.Fid())
}

type Rclunk []byte

func (m Rclunk) Len() int64  { return int64(guint32(m[:4])) }
func (m Rclunk) Tag() uint16 { return guint16(m[5:7]) }
func (m Rclunk) Reset()      {}
func (m Rclunk) String() string {
	return fmt.Sprintf("Rclunk tag:%d", m.Tag())
}

type Tremove []byte

func (m Tremove) Fid() uint32 { return guint32(m[7:11]) }

func (m Tremove) Len() int64  { return int64(guint32(m[:4])) }
func (m Tremove) Tag() uint16 { return guint16(m[5:7]) }
func (m Tremove) Reset()      {}
func (m Tremove) String() string {
	return fmt.Sprintf("Tremove tag:%d fid:%d", m.Tag(), m.Fid())
}

type Rremove []byte

func (m Rremove) Len() int64  { return int64(guint32(m[:4])) }
func (m Rremove) Tag() uint16 { return guint16(m[5:7]) }
func (m Rremove) Reset()      {}
func (m Rremove) String() string {
	return fmt.Sprintf("Rremove tag:%d", m.Tag())
}

type Tstat []byte

func (m Tstat) Fid() uint32 { return guint32(m[7:11]) }

func (m Tstat) Len() int64  { return int64(guint32(m[:4])) }
func (m Tstat) Tag() uint16 { return guint16(m[5:7]) }
func (m Tstat) Reset()      {}
func (m Tstat) String() string {
	return fmt.Sprintf("Tstat tag:%d fid:%d", m.Tag(), m.Fid())
}

type Rstat []byte

func parseRstat(data []byte) (Rstat, error) {
	n := headerLen + 2 + 2 + fixedStatLen - (2 * 4)
	if err := vfield(data, n, 0, verifyName); err != nil {
		return nil, err
	}
	if err := vfield(data, n, 1, verifyUname); err != nil {
		return nil, err
	}
	if err := vfield(data, n, 2, verifyUname); err != nil {
		return nil, err
	}
	if err := vfield(data, n, 3, verifyUname); err != nil {
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
func (m Rstat) Reset()      {}
func (m Rstat) String() string {
	return fmt.Sprintf("Rstat tag:%d stat:%q", m.Tag(), m.Stat())
}

type Twstat []byte

func parseTwstat(data []byte) (Twstat, error) {
	n := headerLen + 4 + 2 + 2 + fixedStatLen - (2 * 4)
	if err := vfield(data, n, 0, verifyName); err != nil {
		return nil, err
	}
	if err := vfield(data, n, 1, verifyUname); err != nil {
		return nil, err
	}
	if err := vfield(data, n, 2, verifyUname); err != nil {
		return nil, err
	}
	if err := vfield(data, n, 3, verifyUname); err != nil {
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
func (m Twstat) Reset()      {}
func (m Twstat) String() string {
	return fmt.Sprintf("Twstat tag:%d fid:%d", m.Tag(), m.Fid())
}

type Rwstat []byte

func (m Rwstat) Len() int64  { return int64(guint32(m[:4])) }
func (m Rwstat) Tag() uint16 { return guint16(m[5:7]) }
func (m Rwstat) Reset()      {}
func (m Rwstat) String() string {
	return fmt.Sprintf("Rwstat tag:%d", m.Tag())
}

type Rerror []byte

func (m Rerror) Ename() string { return string(gfield(m, 7, 0)) }

func (m Rerror) Len() int64  { return int64(guint32(m[:4])) }
func (m Rerror) Tag() uint16 { return guint16(m[5:7]) }
func (m Rerror) Reset()      {}
func (m Rerror) String() string {
	return fmt.Sprintf("Rerror tag:%d ename:%q",
		m.Tag(), gfield(m, 7, 0))
}
