package ninep

import (
	"sync"
	"sync/atomic"

	"github.com/azmodb/ninep/posix"
)

// fid represents a filesystem node and tracks references. The fid will be
// closed only when references reaches zero.
type fid struct {
	// ref is an active reference count
	ref int64

	*posix.Fid
}

func newFid(f *posix.Fid) *fid { return &fid{ref: 0, Fid: f} }

// DecRef should be called once finished with a fid.
func (f *fid) DecRef() {
	if atomic.AddInt64(&f.ref, -1) == 0 {
		if f.Fid != nil {
			f.Fid.Close()
		}
	}
}

func (f *fid) incRef() { atomic.AddInt64(&f.ref, 1) }

// fidmap provide management of fid structures.
type fidmap struct {
	mu sync.Mutex // protects following
	m  map[uint32]*fid
}

// newFidmap initializes and returns a fidmap.
func newFidmap() *fidmap {
	return &fidmap{m: make(map[uint32]*fid)}
}

// Load finds the given fid. DecRef should be called once finished with a
// fid.
func (m *fidmap) Load(num uint32) (*fid, bool) {
	m.mu.Lock()
	f, found := m.m[num]
	m.mu.Unlock()

	if found {
		f.incRef()
	}
	return f, found
}

// Attach inserts the given fid. An error is returned if fid is already
// in use.
func (m *fidmap) Attach(num uint32, f *fid) bool {
	f.incRef()

	m.mu.Lock()
	if _, found := m.m[num]; found {
		m.mu.Unlock()
		return false
	}
	m.m[num] = f
	m.mu.Unlock()

	return true
}

// Store inserts the given fid. This fid starts with a reference count of
// one. If a fid exists in the slot already it is closed, per the
// specification.
func (m *fidmap) Store(num uint32, f *fid) bool {
	f.incRef()

	m.mu.Lock()
	fid, found := m.m[num]
	m.m[num] = f
	m.mu.Unlock()

	if found {
		fid.DecRef()
	}
	return found
}

// Delete removes the given fid and drops a reference.
func (m *fidmap) Delete(num uint32) bool {
	m.mu.Lock()
	fid, found := m.m[num]
	delete(m.m, num)
	m.mu.Unlock()

	if found {
		fid.DecRef()
	}
	return found
}

// Clear resets fidmap.
func (m *fidmap) Clear() {
	m.mu.Lock()
	for num, fid := range m.m {
		delete(m.m, num)
		fid.DecRef()
	}
	m.mu.Unlock()
}
