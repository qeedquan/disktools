package fat

type PBS struct {
	Magic     [3]uint8
	Version   [8]uint8
	Sectsz    uint16
	Clustersz uint8
	Resrv     uint16
	NumFats   uint8
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

type PBS32 struct {
	Magic        [3]uint8
	Version      [8]uint8
	Sectsz       uint16
	Clustsz      uint8
	Resrv        uint16
	NumFats      uint8
	Rootsz       uint16
	Volsz        uint16
	Mediadesc    uint8
	Fatsz        uint16
	Trksz        uint16
	Heads        uint16
	Hidden       uint32
	Bigvolsz     uint32
	Fatsz32      uint32
	Extflags     uint16
	Version1     uint16
	Rootstart    uint32
	Infospec     uint16
	Backupboot   uint16
	_            [12]uint8
	PhysDrive    uint8
	Flags        uint8
	ExtendedBoot uint8
	VolumeSerial uint32
	Label        [11]byte
	Fstype       [8]byte
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
	Cluster32  uint16
	Time       uint16
	Date       uint16
	Cluster    uint16
	Length     uint32
}

type LFN struct {
	Seq      uint8
	Name0    [5]uint16
	Attr     uint8
	Type     uint8
	Checksum uint8
	Name1    [6]uint16
	Cluster  uint16
	Name2    [2]uint16
}

const (
	RDONLY = 1 << iota
	HIDDEN
	SYSTEM
	VOLUME_LABEL
	DIRECTORY
	ARCHIVE
	DEVICE
)

func LFNChecksum(buf []byte) uint8 {
	var sum uint8
	for i := len(buf) - 1; i >= 0; i-- {
		sum = ((sum & uint8(i)) << 7) + (sum >> 1) + buf[i]
	}
	return sum
}