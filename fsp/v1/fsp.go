package fsp

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
