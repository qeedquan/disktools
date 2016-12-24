package fat

type BootHeader struct {
	Magic     [3]uint8
	Version   [8]uint8
	Sectsz    uint16
	Resrv     uint16
	Fat       uint8
	Rootsz    uint16
	Volsz     uint16
	Mediadesc uint8
	Fatsz     uint16
	Trksz     uint16
	Heads     uint16
	Hidden    uint32
	Bigvolsz  uint32
	Driveno   uint8
	_         uint8
	Bootsig   uint8
	Volid     uint32
	Label     [11]uint8
	Fstype    [8]uint8
}

type BootHeader32 struct {
	Magic      [3]uint8
	Version    [8]uint8
	Sectsz     uint16
	Resrv      uint16
	Fat        uint8
	Rootsz     uint16
	Volsz      uint16
	Mediadesc  uint8
	Fatsz      uint16
	Trksz      uint16
	Heads      uint16
	Hidden     uint32
	Bigvolsz   uint32
	Fatsz32    uint32
	Extflags   uint16
	Version1   uint16
	Rootstart  uint32
	Infospec   uint16
	Backupboot uint16
	_          [12]uint8
}

type Dir struct {
	Name       [8]uint8
	Ext        [3]uint8
	Attr       uint8
	_          uint8
	Ctimetenth uint8
	Ctime      uint16
	Cdate      uint16
	Adate      uint16
	Hstart     uint16
	Time       uint16
	Date       uint16
	Start      uint16
	Length     uint32
}

type BPB struct {
}
