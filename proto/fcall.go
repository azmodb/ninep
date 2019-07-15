package proto

// Fcall represents an active 9P RPC.
type Fcall struct {
	// The argument to the 9P function and used to determine the message
	// type.
	Tx Message

	Rx Message // The reply from the function.

	Err error
	C   chan *Fcall
}

// Alloc returns a new fcall object from the internal pool. Alloc is safe
// for use by multiple goroutines simultaneously.
func Alloc(t MessageType) (*Fcall, bool) {
	f, found := allocator.Get(t)
	if !found {
		return nil, false
	}
	return f.(*Fcall), true
}

// Reset resets all state.
func (f *Fcall) Reset() {
	if f.Tx != nil {
		f.Tx.Reset()
	}
	if f.Rx != nil {
		f.Rx.Reset()
	}
	f.C = nil
	f.Err = nil
}

// Release resets all state and adds all fcalls to their pool.
func Release(fcall ...*Fcall) {
	for _, f := range fcall {
		if f.Tx == nil {
			continue
		}
		allocator.Put(f.Tx.MessageType(), f)
	}
}
