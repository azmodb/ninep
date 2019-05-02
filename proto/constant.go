package proto

import "fmt"

// MessageType represents a 9P message type identifier.
type MessageType uint8

// Defines message type identifiers.
const (
	MessageTversion MessageType = 100 + iota
	MessageRversion
	MessageTauth
	MessageRauth
	MessageTattach
	MessageRattach
	MessageTerror // illegal
	MessageRerror
	MessageTflush
	MessageRflush
	MessageTwalk
	MessageRwalk
	MessageTopen
	MessageRopen
	MessageTcreate
	MessageRcreate
	MessageTread
	MessageRread
	MessageTwrite
	MessageRwrite
	MessageTclunk
	MessageRclunk
	MessageTremove
	MessageRremove
	MessageTstat
	MessageRstat
	MessageTwstat
	MessageRwstat

	MessageTlauth   MessageType = 102
	MessageRlauth   MessageType = 103
	MessageTlattach MessageType = 104
	MessageRlattach MessageType = 105

	MessageRlerror      MessageType = 7
	MessageTstatfs      MessageType = 8
	MessageRstatfs      MessageType = 9
	MessageTlopen       MessageType = 12
	MessageRlopen       MessageType = 13
	MessageTlcreate     MessageType = 14
	MessageRlcreate     MessageType = 15
	MessageTsymlink     MessageType = 16
	MessageRsymlink     MessageType = 17
	MessageTmknod       MessageType = 18
	MessageRmknod       MessageType = 19
	MessageTrename      MessageType = 20
	MessageRrename      MessageType = 21
	MessageTreadlink    MessageType = 22
	MessageRreadlink    MessageType = 23
	MessageTgetattr     MessageType = 24
	MessageRgetattr     MessageType = 25
	MessageTsetattr     MessageType = 26
	MessageRsetattr     MessageType = 27
	MessageTxattrwalk   MessageType = 30
	MessageRxattrwalk   MessageType = 31
	MessageTxattrcreate MessageType = 32
	MessageRxattrcreate MessageType = 33
	MessageTreaddir     MessageType = 40
	MessageRreaddir     MessageType = 41
	MessageTfsync       MessageType = 50
	MessageRfsync       MessageType = 51
	MessageTlock        MessageType = 52
	MessageRlock        MessageType = 53
	MessageTgetlock     MessageType = 54
	MessageRgetlock     MessageType = 55
	MessageTlink        MessageType = 70
	MessageRlink        MessageType = 71
	MessageTmkdir       MessageType = 72
	MessageRmkdir       MessageType = 73
	MessageTrenameat    MessageType = 74
	MessageRrenameat    MessageType = 75
	MessageTunlinkat    MessageType = 76
	MessageRunlinkat    MessageType = 77
)

func (t MessageType) String() string {
	switch t {
	case MessageTversion:
		return "Tversion"
	case MessageRversion:
		return "Rversion"
	// case MessageTauth:
	// 	return "Tauth"
	// case MessageRauth:
	// 	return "Rauth"
	// case MessageTattach:
	// 	return "Tattach"
	// case MessageRattach:
	// 	return "Rattach"
	case MessageRerror:
		return "Rerror"
	case MessageTflush:
		return "Tflush"
	case MessageRflush:
		return "Rflush"
	case MessageTwalk:
		return "Twalk"
	case MessageRwalk:
		return "Rwalk"
	case MessageTopen:
		return "Topen"
	case MessageRopen:
		return "Ropen"
	case MessageTcreate:
		return "Tcreate"
	case MessageRcreate:
		return "Rcreate"
	case MessageTread:
		return "Tread"
	case MessageRread:
		return "Rread"
	case MessageTwrite:
		return "Twrite"
	case MessageRwrite:
		return "Rwrite"
	case MessageTclunk:
		return "Tclunk"
	case MessageRclunk:
		return "Rclunk"
	case MessageTremove:
		return "Tremove"
	case MessageRremove:
		return "Rremove"
	case MessageTstat:
		return "Tstat"
	case MessageRstat:
		return "Rstat"
	case MessageTwstat:
		return "Twstat"
	case MessageRwstat:
		return "Rwstat"

	case MessageTlauth:
		return "Tlauth"
	case MessageRlauth:
		return "Rlauth"
	case MessageTlattach:
		return "Tlattach"
	case MessageRlattach:
		return "Rlattach"

	case MessageTstatfs:
		return "Tstatfs"
	case MessageRstatfs:
		return "Rstatfs"
	case MessageTlopen:
		return "Tlopen"
	case MessageRlopen:
		return "Rlopen"
	case MessageTlcreate:
		return "Tlcreate"
	case MessageRlcreate:
		return "Rlcreate"
	case MessageTsymlink:
		return "Tsymlink"
	case MessageRsymlink:
		return "Rsymlink"
	case MessageTmknod:
		return "Tmknod"
	case MessageRmknod:
		return "Rmknod"
	case MessageTrename:
		return "Trename"
	case MessageRrename:
		return "Rrename"
	case MessageTreadlink:
		return "Treadlink"
	case MessageRreadlink:
		return "Rreadlink"
	case MessageTgetattr:
		return "Tgetattr"
	case MessageRgetattr:
		return "Rgetattr"
	case MessageTsetattr:
		return "Tsetattr"
	case MessageRsetattr:
		return "Rsetattr"
	case MessageTxattrwalk:
		return "Txattrwalk"
	case MessageRxattrwalk:
		return "Rxattrwalk"
	case MessageTxattrcreate:
		return "Txattrcreate"
	case MessageRxattrcreate:
		return "Rxattrcreate"
	case MessageTreaddir:
		return "Treaddir"
	case MessageRreaddir:
		return "Rreaddir"
	case MessageTfsync:
		return "Tfsync"
	case MessageRfsync:
		return "Rfsync"
	case MessageTlock:
		return "Tlock"
	case MessageRlock:
		return "Rlock"
	case MessageTgetlock:
		return "Tgetlock"
	case MessageRgetlock:
		return "Rgetlock"
	case MessageTlink:
		return "Tlink"
	case MessageRlink:
		return "Rlink"
	case MessageTmkdir:
		return "Tmkdir"
	case MessageRmkdir:
		return "Rmkdir"
	case MessageTrenameat:
		return "Trenameat"
	case MessageRrenameat:
		return "Rrenameat"
	case MessageTunlinkat:
		return "Tunlinkat"
	case MessageRunlinkat:
		return "Runlinkat"
	}
	return fmt.Sprintf("unknown message identifier (%d)", t)
}
