package posix

func (fs *unixFS) setfsid(uid, gid uint32) error { return nil }

func (fs *unixFS) resetfsid() {}
