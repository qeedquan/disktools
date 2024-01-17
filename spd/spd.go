package spd

type DDR3 struct {
	Size     uint8
	Used     uint8
	Mem      uint8
	Mod      uint8
	Major    uint8
	Minor    uint8
	Dividend uint16
	Divisor  uint8
	MMID     uint16
	Loc      uint8
	Years    uint8
	Weeks    uint8
	CRC      uint16
	MPN      string
	MRC      string
	DMID     uint16
}
