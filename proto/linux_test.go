// THIS FILE IS AUTOMATICALLY GENERATED by `go run internal/generator.go`.
// EDIT linux.go INSTEAD.

package proto

import "math"

var generatedLinuxPackets = []packet{
	packet{&Tlauth{}, &Tlauth{}},
	packet{&Tlauth{AuthFid: math.MaxUint32, UserName: string16.String(), Path: string16.String(), Uid: math.MaxUint32}, &Tlauth{}},
	packet{&Tlattach{}, &Tlattach{}},
	packet{&Tlattach{Fid: math.MaxUint32, AuthFid: math.MaxUint32, UserName: string16.String(), Path: string16.String(), Uid: math.MaxUint32}, &Tlattach{}},
	packet{&Rlerror{}, &Rlerror{}},
	packet{&Rlerror{Errno: math.MaxUint32}, &Rlerror{}},
	packet{&Tstatfs{}, &Tstatfs{}},
	packet{&Tstatfs{Fid: math.MaxUint32}, &Tstatfs{}},
	packet{&Rstatfs{}, &Rstatfs{}},
	packet{&Rstatfs{Type: math.MaxUint32, BlockSize: math.MaxUint32, Blocks: math.MaxUint64, BlocksFree: math.MaxUint64, BlocksAvailable: math.MaxUint64, Files: math.MaxUint64, FilesFree: math.MaxUint64, FsID: math.MaxUint64, NameLength: math.MaxUint32}, &Rstatfs{}},
	packet{&Tlopen{}, &Tlopen{}},
	packet{&Tlopen{Fid: math.MaxUint32, Flags: math.MaxUint32}, &Tlopen{}},
	packet{&Rlopen{}, &Rlopen{}},
	packet{&Rlopen{Qid: Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64}, Iounit: math.MaxUint32}, &Rlopen{}},
	packet{&Tlcreate{}, &Tlcreate{}},
	packet{&Tlcreate{Fid: math.MaxUint32, Name: string16.String(), Flags: math.MaxUint32, Permission: math.MaxUint32, Gid: math.MaxUint32}, &Tlcreate{}},
	packet{&Rlcreate{}, &Rlcreate{}},
	packet{&Rlcreate{Qid: Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64}, Iounit: math.MaxUint32}, &Rlcreate{}},
	packet{&Tsymlink{}, &Tsymlink{}},
	packet{&Tsymlink{DirectoryFid: math.MaxUint32, Name: string16.String(), Target: string16.String(), Gid: math.MaxUint32}, &Tsymlink{}},
	packet{&Tmknod{}, &Tmknod{}},
	packet{&Tmknod{DirectoryFid: math.MaxUint32, Name: string16.String(), Permission: math.MaxUint32, Major: math.MaxUint32, Minor: math.MaxUint32, Gid: math.MaxUint32}, &Tmknod{}},
	packet{&Trename{}, &Trename{}},
	packet{&Trename{Fid: math.MaxUint32, DirectoryFid: math.MaxUint32, Name: string16.String()}, &Trename{}},
	packet{&Rrename{}, &Rrename{}},
	packet{&Treadlink{}, &Treadlink{}},
	packet{&Treadlink{Fid: math.MaxUint32}, &Treadlink{}},
	packet{&Rreadlink{}, &Rreadlink{}},
	packet{&Rreadlink{Target: string16.String()}, &Rreadlink{}},
	packet{&Tgetattr{}, &Tgetattr{}},
	packet{&Tgetattr{Fid: math.MaxUint32, AttrMask: math.MaxUint64}, &Tgetattr{}},
	packet{&Rgetattr{}, &Rgetattr{}},
	packet{&Rgetattr{Valid: math.MaxUint64, Qid: Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64}, Mode: math.MaxUint32, Uid: math.MaxUint32, Gid: math.MaxUint32, Nlink: math.MaxUint64, Rdev: math.MaxUint64, Size: math.MaxUint64, BlockSize: math.MaxUint64, Blocks: math.MaxUint64, Gen: math.MaxUint64, DataVersion: math.MaxUint64}, &Rgetattr{}},
	packet{&Tsetattr{}, &Tsetattr{}},
	packet{&Tsetattr{Fid: math.MaxUint32, Valid: math.MaxUint32, Mode: math.MaxUint32, Uid: math.MaxUint32, Gid: math.MaxUint32, Size: math.MaxUint64}, &Tsetattr{}},
	packet{&Rsetattr{}, &Rsetattr{}},
	packet{&Txattrwalk{}, &Txattrwalk{}},
	packet{&Txattrwalk{Fid: math.MaxUint32, NewFid: math.MaxUint32, Name: string16.String()}, &Txattrwalk{}},
	packet{&Rxattrwalk{}, &Rxattrwalk{}},
	packet{&Rxattrwalk{Size: math.MaxUint64}, &Rxattrwalk{}},
	packet{&Txattrcreate{}, &Txattrcreate{}},
	packet{&Txattrcreate{Fid: math.MaxUint32, Name: string16.String(), AttrSize: math.MaxUint64, Flag: math.MaxUint32}, &Txattrcreate{}},
	packet{&Rxattrcreate{}, &Rxattrcreate{}},
	packet{&Treaddir{}, &Treaddir{}},
	packet{&Treaddir{Fid: math.MaxUint32, Offset: math.MaxUint64, Count: math.MaxUint32}, &Treaddir{}},
}
