package ninep

type buffer []byte

func (b *buffer) WriteString(s string) (int, error) {
	n := b.grow(len(s))
	return copy((*b)[n:], s), nil
}

func (b *buffer) Write(p []byte) (int, error) {
	n := b.grow(len(p))
	return copy((*b)[n:], p), nil
}

func (b *buffer) WriteByte(p byte) error {
	n := b.grow(1)
	(*b)[n] = p
	return nil
}

func (b *buffer) grow(size int) int {
	n := len(*b)
	switch {
	case n == 0 && size < 256:
		*b = make([]byte, 256)
	case n+size > cap(*b):
		m := n + size*2*cap(*b)
		if m > 2*8192 {
			m = n + size
		}
		nb := make([]byte, n, m)
		copy(nb, *b)
		*b = nb
	}
	*b = (*b)[0 : n+size]
	return n
}
