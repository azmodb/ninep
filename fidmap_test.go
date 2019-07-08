package ninep

import "testing"

func TestFidmapAttachStoreDelete(t *testing.T) {
	m := newFidmap()
	if success := m.Attach(42, newFid(nil)); !success {
		t.Fatalf("attach: unepxected failure")
	}
	if success := m.Attach(42, newFid(nil)); success {
		t.Fatalf("attach: unepxected success")
	}

	if success := m.Store(42, newFid(nil)); !success {
		t.Fatalf("store: unepxected failure")
	}
	if success := m.Store(42, newFid(nil)); !success {
		t.Fatalf("store: unepxected failure")
	}

	if _, success := m.Load(42); !success {
		t.Fatalf("load: unepxected failure")
	}

	if success := m.Delete(42); !success {
		t.Fatalf("delete: unepxected failure")
	}
	if success := m.Delete(42); success {
		t.Fatalf("delete: unepxected success")
	}
}
