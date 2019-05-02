package proto

import "math"

var generatedLegacyPackets = []packet{
	packet{&Header{}, &Header{}},
	packet{&Header{Type: math.MaxUint8, Tag: math.MaxUint16}, &Header{}},
	packet{&Tversion{}, &Tversion{}},
	packet{&Tversion{MessageSize: math.MaxUint32, Version: string16.String()}, &Tversion{}},
	packet{&Rversion{}, &Rversion{}},
	packet{&Rversion{MessageSize: math.MaxUint32, Version: string16.String()}, &Rversion{}},
	packet{&Tflush{}, &Tflush{}},
	packet{&Tflush{OldTag: math.MaxUint16}, &Tflush{}},
	packet{&Rflush{}, &Rflush{}},
	packet{&Tread{}, &Tread{}},
	packet{&Tread{Fid: math.MaxUint32, Offset: math.MaxUint64, Count: math.MaxUint32}, &Tread{}},
	packet{&Rwrite{}, &Rwrite{}},
	packet{&Rwrite{Count: math.MaxUint32}, &Rwrite{}},
	packet{&Tclunk{}, &Tclunk{}},
	packet{&Tclunk{Fid: math.MaxUint32}, &Tclunk{}},
	packet{&Rclunk{}, &Rclunk{}},
	packet{&Tremove{}, &Tremove{}},
	packet{&Tremove{Fid: math.MaxUint32}, &Tremove{}},
	packet{&Rremove{}, &Rremove{}},
}
