package fsp

import "github.com/qeedquan/disktools/efi"

type Header struct {
	Signature              [4]uint8
	HeaderLength           uint32
	_                      [3]uint8
	HeaderRevision         uint8
	ImageRevision          uint32
	ImageId                uint64
	ImageSize              uint32
	ImageBase              uint32
	ImageAttribute         uint32
	CfgRegionOffset        uint32
	CfgRegionSize          uint32
	ApiEntryNum            uint32
	TempRamInitEntryOffset uint32
	FspInitEntryOffset     uint32
	NotifyPhaseEntryOffset uint32
	_                      [4]byte
}

type File struct {
	VolumnHeader    efi.VolumnHeader
	VolumnExtHeader efi.VolumnExtHeader
	FfsHeader       efi.FfsFileHeader
	FspHeader       Header
	VPD             interface{}
	UPD             interface{}
	Firmware        []byte
}

type PlatData struct {
	Data                []byte
	MicrocodeRegionBase uint32
	MicrocodeRegionSize uint32
	CodeRegionBase      uint32
	CodeRegionSize      uint32
}

type GlobalData struct {
	Signature     uint32
	CoreStack     uint32
	PlatData      PlatData
	FspInfoHeader Header
	UpdDataRgnPtr interface{}
	ApiMode       uint8
	_             [3]uint8
	PerfIdx       uint32
	PerfData      [32]uint64
}

type Error int

const (
	SUCCESS              Error = 0x00000000
	INVALID_PARAMETER    Error = 0x80000002
	UNSUPPORTED          Error = 0x80000003
	NOT_READY            Error = 0x80000006
	DEVICE_ERROR         Error = 0x80000007
	OUT_OF_RESOURCES     Error = 0x80000009
	VOLUME_CORRUPTED     Error = 0x8000000A
	NOT_FOUND            Error = 0x8000000E
	TIMEOUT              Error = 0x80000012
	ABORTED              Error = 0x80000015
	INCOMPATIBLE_VERSION Error = 0x80000010
	SECURITY_VIOLATION   Error = 0x8000001A
	CRC_ERROR            Error = 0x8000001B
)

func (e Error) Error() string {
	switch e {
	case SUCCESS:
		return "success"
	case INVALID_PARAMETER:
		return "invalid parameter"
	case UNSUPPORTED:
		return "unsupported"
	case NOT_READY:
		return "not ready"
	case DEVICE_ERROR:
		return "device error"
	case OUT_OF_RESOURCES:
		return "out of resources"
	case VOLUME_CORRUPTED:
		return "volume corrupted"
	case NOT_FOUND:
		return "not found"
	case TIMEOUT:
		return "timeout"
	case ABORTED:
		return "aborted"
	case INCOMPATIBLE_VERSION:
		return "incompatible version"
	case SECURITY_VIOLATION:
		return "security violation"
	case CRC_ERROR:
		return "crc error"
	}
	return "unknown error"
}

const (
	VPD_MIN_SIZE = 0x20
	UPD_MIN_SIZE = 0x20
)
