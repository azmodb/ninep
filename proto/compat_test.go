package proto

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/azmodb/ninep/proto/internal/plan9"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	N           = 8192 // make this bigger for a larger (and slower) test
)

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

const (
	maxUint64 = math.MaxUint64
	maxUint32 = math.MaxUint32
	maxUint16 = math.MaxUint16
	maxUint8  = math.MaxUint8
)

var (
	maxQids     [MaxWalkElements]plan9.Qid
	maxQid      plan9.Qid
	maxElements [MaxWalkElements]string

	maxDir = plan9.Dir{
		Type:  maxUint16,
		Dev:   maxUint32,
		Qid:   maxQid,
		Mode:  maxUint32,
		Atime: maxUint32,
		Mtime: maxUint32,
		Name:  maxDirName,
		Uid:   maxUserName,
		Gid:   maxUserName,
		Muid:  maxUserName,
	}
	maxDirData []byte

	maxVersion  string
	maxUserName string
	maxDirName  string
	maxPath     string
	maxEname    string

	testBytes [N]byte
)

func init() {
	rand.Seed(time.Now().UnixNano())

	maxQid = plan9.Qid{Type: maxUint8, Vers: maxUint32, Path: maxUint64}

	maxUserName = randString(MaxUserNameSize)
	maxDirName = randString(MaxDirNameSize)
	maxPath = randString(MaxPathSize)
	maxEname = randString(MaxEnameSize)

	for i := 0; i < MaxWalkElements; i++ {
		maxElements[i] = randString(MaxDirNameSize)
		maxQids[i] = maxQid
	}

	for i := 0; i < N; i++ {
		testBytes[i] = 'a' + byte(i%26)
	}

	maxDir.Name = maxDirName
	maxDir.Uid = maxUserName
	maxDir.Gid = maxUserName
	maxDir.Muid = maxUserName

	maxDirData, _ = maxDir.Bytes()
}

var testFcalls = []plan9.Fcall{
	//plan9.NewVersion(plan9.Tversion, maxUint16, MaxMessageSize, maxVersion),
	//plan9.NewVersion(plan9.Rversion, maxUint16, MaxMessageSize, maxVersion),

	plan9.NewVersion(plan9.Tversion, 42, 8192, Version),
	plan9.NewVersion(plan9.Rversion, 42, 8192, Version),
	plan9.NewVersion(plan9.Tversion, 0, 0, ""),
	plan9.NewVersion(plan9.Rversion, 0, 0, ""),

	plan9.NewTauth(maxUint16, maxUint32, maxUserName, maxPath),
	plan9.NewTauth(42, 42, "bootes", "/usr/bootes"),
	plan9.NewTauth(0, 0, "", ""),

	plan9.NewTattach(maxUint16, maxUint32, maxUint32, maxUserName, maxPath),
	plan9.NewTattach(42, 42, 43, "bootes", "/usr/bootes"),
	plan9.NewTattach(0, 0, 0, "", ""),

	plan9.NewRauth(maxUint16, maxQid),
	plan9.NewRauth(0, plan9.Qid{}),
	plan9.NewRattach(maxUint16, maxQid),
	plan9.NewRattach(0, plan9.Qid{}),

	plan9.NewRerror(maxUint16, maxEname),
	plan9.NewRerror(42, "some error"),
	plan9.NewRerror(0, ""),

	plan9.NewTflush(maxUint16, maxUint16),
	plan9.NewTflush(42, 43),
	plan9.NewTflush(0, 0),
	plan9.NewRflush(maxUint16),
	plan9.NewRflush(42),
	plan9.NewRflush(0),

	plan9.NewTwalk(maxUint16, maxUint32, maxUint32, maxElements[:]),
	plan9.NewTwalk(42, 20, 21, maxElements[:2]),
	plan9.NewTwalk(0, 0, 0, nil),

	plan9.NewRwalk(maxUint16, maxQids[:]),
	plan9.NewRwalk(42, maxQids[:3]),
	plan9.NewRwalk(0, nil),

	plan9.NewTopen(maxUint16, maxUint32, maxUint8),
	plan9.NewTopen(42, 43, 44),
	plan9.NewTopen(0, 0, 0),

	plan9.NewRopen(maxUint16, maxQid, maxUint32),
	plan9.NewRopen(42, plan9.Qid{Vers: 42}, 44),
	plan9.NewRopen(0, plan9.Qid{}, 0),

	plan9.NewTcreate(maxUint16, maxUint32, maxDirName, maxUint32, maxUint8),
	plan9.NewTcreate(42, 43, "newfile", 44, 45),
	plan9.NewTcreate(0, 0, "", 0, 0),

	plan9.NewRcreate(maxUint16, maxQid, maxUint32),
	plan9.NewRcreate(42, plan9.Qid{Vers: 42}, 44),
	plan9.NewRcreate(0, plan9.Qid{}, 0),

	plan9.NewTread(maxUint16, maxUint32, maxUint64, maxUint32),
	plan9.NewTread(0, 0, 0, 0),

	plan9.NewRread(maxUint16, testBytes[:]),
	plan9.NewRread(42, testBytes[:26]),
	plan9.NewRread(0, nil),

	plan9.NewTwrite(maxUint16, maxUint32, maxUint64, testBytes[:]),
	plan9.NewTwrite(42, 43, 44, testBytes[:26]),
	plan9.NewTwrite(0, 0, 0, nil),

	plan9.NewRwrite(maxUint16, maxUint32),
	plan9.NewRwrite(0, 0),

	plan9.NewTclunk(maxUint16, maxUint32),
	plan9.NewTclunk(42, 43),
	plan9.NewTclunk(0, 0),

	plan9.NewRclunk(maxUint16),
	plan9.NewRclunk(42),
	plan9.NewRclunk(0),

	plan9.NewTremove(maxUint16, maxUint32),
	plan9.NewTremove(42, 43),
	plan9.NewTremove(0, 0),

	plan9.NewRremove(maxUint16),
	plan9.NewRremove(42),
	plan9.NewRremove(0),

	plan9.NewTstat(maxUint16, maxUint32),
	plan9.NewTstat(42, 43),
	plan9.NewTstat(0, 0),

	plan9.NewRstat(maxUint16, maxDirData),
	plan9.NewRstat(0, nil),

	plan9.NewTwstat(maxUint16, maxUint32, maxDirData),
	plan9.NewTwstat(0, 0, nil),

	plan9.NewRwstat(maxUint16),
	plan9.NewRwstat(42),
	plan9.NewRwstat(0),
}

func compareQid(a Qid, b plan9.Qid) error {
	if a.Type != b.Type {
		return fmt.Errorf("expected qid type (%d), got (%d)", b.Type, a.Type)
	}
	if a.Version != b.Vers {
		return fmt.Errorf("expected qid version (%d), got (%d)", b.Vers, a.Version)
	}
	if a.Path != b.Path {
		return fmt.Errorf("expected qid path (%d), got (%d)", b.Path, a.Path)
	}
	return nil
}

func compareStat(stat Stat, dir plan9.Dir) error {
	if stat.Type != dir.Type {
		return fmt.Errorf("expected type (%d), got (%d)", dir.Type, stat.Type)
	}
	if stat.Dev != dir.Dev {
		return fmt.Errorf("expected dev (%d), got (%d)", dir.Dev, stat.Dev)
	}
	if err := compareQid(stat.Qid, dir.Qid); err != nil {
		return err
	}
	if stat.Mode != uint32(dir.Mode) {
		return fmt.Errorf("expected perm (%d), got (%d)", dir.Mode, stat.Mode)
	}
	if stat.Atime != dir.Atime {
		return fmt.Errorf("expected atime (%d), got (%d)", dir.Atime, stat.Atime)
	}
	if stat.Mtime != dir.Mtime {
		return fmt.Errorf("expected mtime (%d), got (%d)", dir.Mtime, stat.Mtime)
	}
	if stat.Name != dir.Name {
		return fmt.Errorf("expected name %q, got %q", dir.Name, stat.Name)
	}
	if stat.Uid != dir.Uid {
		return fmt.Errorf("expected uid %q, got %q", dir.Uid, stat.Uid)
	}
	if stat.Gid != dir.Gid {
		return fmt.Errorf("expected gid %q, got %q", dir.Gid, stat.Gid)
	}
	if stat.Muid != dir.Muid {
		return fmt.Errorf("expected muid %q, got %q", dir.Muid, stat.Muid)
	}
	return nil
}

func TestStatEncodingCompatiblity(t *testing.T) {
	for i, dir := range []plan9.Dir{plan9.Dir{}, maxDir} {
		dirBytes, err := dir.Bytes()
		if err != nil {
			t.Errorf("stat test[#%d]: cannot marshal plan9 stat: %v", i, err)
		}

		stat := Stat{}
		if err := stat.UnmarshalBinary(dirBytes); err != nil {
			t.Errorf("stat test[#%d]: cannot unmarshal stat: %v", i, err)
		}

		if err = compareStat(stat, dir); err != nil {
			t.Errorf("stat test[#%d]: stat differ: %v", i, err)
		}

		statBytes, err := stat.MarshalBinary()
		if err != nil {
			t.Errorf("stat test[#%d]: cannot marshal stat: %v", i, err)
		}

		if !bytes.Equal(statBytes, dirBytes) {
			t.Errorf("stat test[#%d]: differ:\n\t%q\n\t%q", i, dirBytes, statBytes)
		}
	}
}

func qidToPlan9Qid(qid Qid) plan9.Qid {
	return plan9.Qid{Type: qid.Type, Vers: qid.Version, Path: qid.Path}
}

func qidsToPlan9Qids(qids []Qid) []plan9.Qid {
	if len(qids) == 0 {
		return nil
	}

	v := make([]plan9.Qid, 0, len(qids))
	for _, qid := range qids {
		v = append(v, qidToPlan9Qid(qid))
	}
	return v
}

func plan9QidToQid(qid plan9.Qid) Qid {
	return Qid{Type: qid.Type, Version: qid.Vers, Path: qid.Path}
}

func plan9QidsToQids(qids []plan9.Qid) []Qid {
	if len(qids) == 0 {
		return nil
	}

	v := make([]Qid, 0, len(qids))
	for _, qid := range qids {
		v = append(v, plan9QidToQid(qid))
	}
	return v
}

func encodePlan9Fcall(t *testing.T, enc *Encoder, f *plan9.Fcall) (err error) {
	t.Helper()

	switch f.Type {
	case plan9.Tversion:
		err = enc.Tversion(f.Tag, int64(f.Msize), f.Version)
	case plan9.Rversion:
		err = enc.Rversion(f.Tag, int64(f.Msize), f.Version)
	case plan9.Tauth:
		err = enc.Tauth(f.Tag, f.Afid, f.Uname, f.Aname)
	case plan9.Tattach:
		err = enc.Tattach(f.Tag, f.Fid, f.Afid, f.Uname, f.Aname)
	case plan9.Rauth:
		err = enc.Rauth(f.Tag, f.Aqid.Type, f.Aqid.Vers, f.Aqid.Path)
	case plan9.Rattach:
		err = enc.Rattach(f.Tag, f.Qid.Type, f.Qid.Vers, f.Qid.Path)
	case plan9.Rerror:
		err = enc.Rerror(f.Tag, f.Ename)
	case plan9.Tflush:
		err = enc.Tflush(f.Tag, f.Oldtag)
	case plan9.Rflush:
		err = enc.Rflush(f.Tag)
	case plan9.Twalk:
		err = enc.Twalk(f.Tag, f.Fid, f.Newfid, f.Wname...)
	case plan9.Rwalk:
		err = enc.Rwalk(f.Tag, plan9QidsToQids(f.Wqid)...)
	case plan9.Topen:
		err = enc.Topen(f.Tag, f.Fid, f.Mode)
	case plan9.Ropen:
		err = enc.Ropen(f.Tag, f.Qid.Type, f.Qid.Vers, f.Qid.Path, f.Iounit)
	case plan9.Tcreate:
		err = enc.Tcreate(f.Tag, f.Fid, f.Name, uint32(f.Perm), f.Mode)
	case plan9.Rcreate:
		err = enc.Rcreate(f.Tag, f.Qid.Type, f.Qid.Vers, f.Qid.Path, f.Iounit)
	case plan9.Tread:
		err = enc.Tread(f.Tag, f.Fid, f.Offset, f.Count)
	case plan9.Rread:
		err = enc.Rread(f.Tag, f.Data)
	case plan9.Twrite:
		err = enc.Twrite(f.Tag, f.Fid, f.Offset, f.Data)
	case plan9.Rwrite:
		err = enc.Rwrite(f.Tag, f.Count)
	case plan9.Tclunk:
		err = enc.Tclunk(f.Tag, f.Fid)
	case plan9.Rclunk:
		err = enc.Rclunk(f.Tag)
	case plan9.Tremove:
		err = enc.Tremove(f.Tag, f.Fid)
	case plan9.Rremove:
		err = enc.Rremove(f.Tag)
	case plan9.Tstat:
		err = enc.Tstat(f.Tag, f.Fid)
	case plan9.Rstat:
		err = enc.Rstat(f.Tag, f.Stat)
	case plan9.Twstat:
		err = enc.Twstat(f.Tag, f.Fid, f.Stat)
	case plan9.Rwstat:
		err = enc.Rwstat(f.Tag)
	}

	return err
}

func TestEncoderCompatiblity(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)

	for i, f := range testFcalls {
		buf.Reset()

		data, err := f.Bytes()
		if err != nil {
			t.Errorf("encode test[#%d]: cannot marshal plan9 fcall: %v", i, err)
		}

		err = encodePlan9Fcall(t, enc, &f)
		if err != nil {
			t.Errorf("encode test[#%d]: cannot encode plan9 fcall: %v", i, err)
		}

		if !bytes.Equal(data, buf.Bytes()) {
			t.Errorf("encode test[#%d]: differ:\n\t%q\n\t%q", i, data, buf.Bytes())
		}
	}
}

func decodePlan9Fcall(buf *FcallBuffer, tag uint16, f, v *plan9.Fcall) error {
	switch f.Type {
	case plan9.Tversion:
		msize, version := buf.Tversion()
		*v = plan9.NewVersion(plan9.Tversion, tag, uint32(msize), version)
	case plan9.Rversion:
		msize, version := buf.Tversion()
		*v = plan9.NewVersion(plan9.Rversion, tag, uint32(msize), version)
	case plan9.Tauth:
		afid, uname, aname := buf.Tauth()
		*v = plan9.NewTauth(tag, afid, uname, aname)
	case plan9.Tattach:
		fid, afid, uname, aname := buf.Tattach()
		*v = plan9.NewTattach(tag, fid, afid, uname, aname)
	case plan9.Rauth:
		typ, version, path := buf.Rauth()
		*v = plan9.NewRauth(tag, plan9.Qid{path, version, typ})
	case plan9.Rattach:
		typ, version, path := buf.Rattach()
		*v = plan9.NewRattach(tag, plan9.Qid{path, version, typ})
	case plan9.Rerror:
		*v = plan9.NewRerror(tag, buf.Rerror())
	case plan9.Tflush:
		*v = plan9.NewTflush(tag, buf.Tflush())
	case plan9.Rflush:
		*v = plan9.NewRflush(tag)
	case plan9.Twalk:
		fid, newfid, names := buf.Twalk()
		*v = plan9.NewTwalk(tag, fid, newfid, names)
	case plan9.Rwalk:
		qids := buf.Rwalk()
		*v = plan9.NewRwalk(tag, qidsToPlan9Qids(qids))
	case plan9.Topen:
		fid, mode := buf.Topen()
		*v = plan9.NewTopen(tag, fid, mode)
	case plan9.Ropen:
		typ, version, path, iounit := buf.Ropen()
		*v = plan9.NewRopen(tag, plan9.Qid{path, version, typ}, iounit)
	case plan9.Tcreate:
		fid, name, perm, mode := buf.Tcreate()
		*v = plan9.NewTcreate(tag, fid, name, perm, mode)
	case plan9.Rcreate:
		typ, version, path, iounit := buf.Rcreate()
		*v = plan9.NewRcreate(tag, plan9.Qid{path, version, typ}, iounit)
	case plan9.Tread:
		fid, offset, count := buf.Tread()
		*v = plan9.NewTread(tag, fid, offset, count)
	case plan9.Rread:
		*v = plan9.NewRread(tag, buf.Rread())
	case plan9.Twrite:
		fid, offset, data := buf.Twrite()
		*v = plan9.NewTwrite(tag, fid, offset, data)
	case plan9.Rwrite:
		*v = plan9.NewRwrite(tag, buf.Rwrite())
	case plan9.Tclunk:
		*v = plan9.NewTclunk(tag, buf.Tclunk())
	case plan9.Rclunk:
		*v = plan9.NewRclunk(tag)
	case plan9.Tremove:
		*v = plan9.NewTremove(tag, buf.Tremove())
	case plan9.Rremove:
		*v = plan9.NewRremove(tag)
	case plan9.Tstat:
		*v = plan9.NewTstat(tag, buf.Tstat())
	case plan9.Rstat:
		*v = plan9.NewRstat(tag, buf.Rstat())
	case plan9.Twstat:
		fid, data := buf.Twstat()
		*v = plan9.NewTwstat(tag, fid, data)
	case plan9.Rwstat:
		*v = plan9.NewRwstat(tag)
	}
	return buf.Err()
}

func TestDecoderCompatiblity(t *testing.T) {
	buf := &bytes.Buffer{}
	dec := NewDecoder(buf)
	v := plan9.Fcall{}

	for i, f := range testFcalls {
		buf.Reset()

		data, err := f.Bytes()
		if err != nil {
			t.Errorf("decode test[#%d]: cannot marshal plan9 fcall: %v", i, err)
		}

		buf.Write(data)
		want := buf.Len() - 7

		typ, tag, b, err := dec.Next()
		if err != nil {
			t.Errorf("decode test[#%d]: cannot decode plan9 fcall: %v", i, err)
		}

		if uint8(typ) != f.Type {
			t.Errorf("decode test[#%d]: expected type %d, got %d", i, f.Type, typ)
		}
		if tag != f.Tag {
			t.Errorf("decode test[#%d]: expected tag %d, got %d", i, f.Tag, tag)
		}

		if b.Len() != want {
			t.Errorf("decode test[#%d]: expected size %d, got %d", i, want, b.Len())
		}

		v = plan9.Fcall{}
		if err := decodePlan9Fcall(b, tag, &f, &v); err != nil {
			t.Errorf("decode test[#%d]: cannot unmarshal fcall: %v", i, err)
		}
		if !reflect.DeepEqual(v, f) {
			t.Errorf("decode test[#%d]: differ:\n\t%v\n\t%v", i, f, v)
		}
	}
}
