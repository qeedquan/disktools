package gpt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/qeedquan/disktools/endian"
	"github.com/qeedquan/disktools/mbr"
)

type Option struct {
	Sectsz int
}

type GUID [16]byte

type Header struct {
	Sig     [8]byte
	Rev     uint32
	Hdrsz   uint32
	Hdrcrc  uint32
	_       uint32
	Current uint64
	Backup  uint64
	First   uint64
	Last    uint64
	GUID    GUID
	Table   uint64
	Ent     uint32
	Entsz   uint32
	Tabcrc  uint32
}

type Entry struct {
	Part  GUID
	Uniq  GUID
	First uint64
	Last  uint64
	Attr  uint64
	Name  [72]byte
}

type Table struct {
	MBR     *mbr.Record
	Header  Header
	Sectsz  int
	Entries []Entry
}

var (
	ErrHeader = errors.New("gpt: invalid header")
)

func Open(r io.ReaderAt, o *Option) (*Table, error) {
	if o == nil {
		o = &Option{Sectsz: 512}
	}
	d := decoder{
		r:     r,
		Table: Table{Sectsz: o.Sectsz},
	}
	err := d.decode()
	if err != nil {
		return nil, err
	}
	return &d.Table, nil
}

type decoder struct {
	Table
	r io.ReaderAt
}

func (d *decoder) decode() error {
	var err error

	d.MBR, err = mbr.Open(d.r)
	if err != nil {
		return err
	}

	if d.MBR.Part[0].Type != 0xee {
		return ErrHeader
	}

	d.Header, err = d.readHeader(int64(d.Sectsz))
	if err != nil {
		return err
	}

	d.Entries, err = d.readEntry(int64(d.Sectsz * 2))
	if err != nil {
		return err
	}

	return nil
}

func (d *decoder) readHeader(off int64) (Header, error) {
	var h Header
	sr := io.NewSectionReader(d.r, off, math.MaxUint32)
	err := binary.Read(sr, binary.LittleEndian, &h)
	if err != nil {
		return h, err
	}

	if string(h.Sig[:]) != "EFI PART" {
		return h, ErrHeader
	}

	return h, nil
}

func (d *decoder) readEntry(off int64) ([]Entry, error) {
	var entries []Entry

	h := &d.Header
	buf := make([]byte, h.Ent)
	for i := uint32(0); i < h.Ent; i++ {
		_, err := d.r.ReadAt(buf, off)
		if err != nil {
			return nil, err
		}

		var entry Entry
		rd := bytes.NewReader(buf)
		err = binary.Read(rd, binary.LittleEndian, &entry)
		entries = append(entries, entry)
	}

	return entries, nil
}

func ParseGUID(guid string) ([16]byte, error) {
	var (
		a       uint32
		b, c, d uint16
		e       uint64
		p       [16]byte
	)
	n, err := fmt.Sscanf(guid, "%x-%x-%x-%x-%x", &a, &b, &c, &d, &e)
	if err != nil {
		return p, err
	}
	if n != 5 {
		return p, errors.New("invalid GUID format")
	}

	endian.Put32le(p[0:], a)
	endian.Put16le(p[4:], b)
	endian.Put16le(p[6:], c)
	endian.Put16le(p[8:], d)
	endian.Put48le(p[10:], e)

	return p, nil
}

func MustParseGUID(guid string) GUID {
	p, err := ParseGUID(guid)
	if err != nil {
		panic(err)
	}
	return p
}

func (p GUID) String() string {
	return fmt.Sprintf("%X-%X-%X-%X-%X",
		endian.Read32be(p[0:]),
		endian.Read32be(p[4:]),
		endian.Read32be(p[6:]),
		endian.Read32be(p[8:]),
		endian.Read48be(p[10:]),
	)
}

var Parts = []struct {
	Name string
	Desc string
	GUID GUID
}{
	{"unused", "Unused entry", MustParseGUID("00000000-0000-0000-0000-000000000000")},
	{"mbr", "MBR", MustParseGUID("024DEE41-33E7-11D3-9D69-0008C781F39F")},
	{"efi", "EFI System", MustParseGUID("C12A7328-F81F-11D2-BA4B-00A0C93EC93B")},
	{"bios", "BIOS Boot", MustParseGUID("21686148-6449-6E6F-744E-656564454649")},
	{"iffs", "Intel Fast Flash", MustParseGUID("D3BFE2DE-3DAF-11DF-BA40-E3A556D89593")},
	{"sony", "Sony boot", MustParseGUID("F4019732-066E-4E12-8273-346C5641494F")},
	{"lenovo", "Lenovo boot", MustParseGUID("BFBFAFE7-A34F-448A-9A5B-6213EB736C22")},
	{"msr", "Microsoft Reserved", MustParseGUID("E3C9E316-0B5C-4DB8-817D-F92DF00215AE")},
	{"dos", "Microsoft Basic data", MustParseGUID("EBD0A0A2-B9E5-4433-87C0-68B6B72699C7")},
	{"ldmm", "Microsoft Logical Disk Manager metadata", MustParseGUID("5808C8AA-7E8F-42E0-85D2-E1E90434CFB3")},
	{"ldmd", "Microsoft Logical Disk Manager data", MustParseGUID("AF9B60A0-1431-4F62-BC68-3311714A69AD")},
	{"recovery", "Windows Recovery Environment", MustParseGUID("DE94BBA4-06D1-4D40-A16A-BFD50179D6AC")},
	{"gpfs", "IBM General Parallel File System", MustParseGUID("37AFFC90-EF7D-4E96-91C3-2D7AE055B174")},
	{"storagespaces", "Storage Spaces", MustParseGUID("E75CAF8F-F680-4CEE-AFA3-B001E56EFC2D")},
	{"hpuxdata", "HP-UX Data", MustParseGUID("75894C1E-3AEB-11D3-B7C1-7B03A0000000")},
	{"hpuxserv", "HP-UX Service", MustParseGUID("E2A1E728-32E3-11D6-A682-7B03A0000000")},
	{"linuxdata", "Linux Data", MustParseGUID("0FC63DAF-8483-4772-8E79-3D69D8477DE4")},
	{"linuxraid", "Linux RAID", MustParseGUID("A19D880F-05FC-4D3B-A006-743F0F84911E")},
	{"linuxrootx86", "Linux Root (x86)", MustParseGUID("44479540-F297-41B2-9AF7-D131D5F0458A")},
	{"linuxrootx86_64", "Linux Root (x86-64)", MustParseGUID("4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709")},
	{"linuxrootarm", "Linux Root (ARM)", MustParseGUID("69DAD710-2CE4-4E3C-B16C-21A1D49ABED3")},
	{"linuxrootaarch64", "Linux Root (ARM)", MustParseGUID("B921B045-1DF0-41C3-AF44-4C6F280D3FAE")},
	{"linuxswap", "Linux Swap", MustParseGUID("0657FD6D-A4AB-43C4-84E5-0933C84B4F4F")},
	{"linuxlvm", "Linux Logical Volume Manager", MustParseGUID("E6D6D379-F507-44C2-A23C-238F2A3DF928")},
	{"linuxhome", "Linux /home", MustParseGUID("933AC7E1-2EB4-4F13-B844-0E14E2AEF915")},
	{"linuxsrv", "Linux /srv", MustParseGUID("3B8F8425-20E0-4F3B-907F-1A25A76F98E8")},
	{"linuxcrypt", "Linux Plain dm-crypt", MustParseGUID("7FFEC5C9-2D00-49B7-8941-3EA10A5586B7")},
	{"luks", "LUKS", MustParseGUID("CA7D7CCB-63ED-4C53-861C-1742536059CC")},
	{"linuxreserved", "Linux Reserved", MustParseGUID("8DA63339-0007-60C0-C436-083AC8230908")},
	{"fbsdboot", "FreeBSD Boot", MustParseGUID("83BD6B9D-7F41-11DC-BE0B-001560B84F0F")},
	{"fbsddata", "FreeBSD Data", MustParseGUID("516E7CB4-6ECF-11D6-8FF8-00022D09712B")},
	{"fbsdswap", "FreeBSD Swap", MustParseGUID("516E7CB5-6ECF-11D6-8FF8-00022D09712B")},
	{"fbsdufs", "FreeBSD Unix File System", MustParseGUID("516E7CB6-6ECF-11D6-8FF8-00022D09712B")},
	{"fbsdvvm", "FreeBSD Vinum volume manager", MustParseGUID("516E7CB8-6ECF-11D6-8FF8-00022D09712B")},
	{"fbsdzfs", "FreeBSD ZFS", MustParseGUID("516E7CBA-6ECF-11D6-8FF8-00022D09712B")},
	{"applehfs", "Apple HFS+", MustParseGUID("48465300-0000-11AA-AA11-00306543ECAC")},
	{"appleufs", "Apple UFS", MustParseGUID("55465300-0000-11AA-AA11-00306543ECAC")},
	{"applezfs", "Apple ZFS", MustParseGUID("6A898CC3-1DD2-11B2-99A6-080020736631")},
	{"appleraid", "Apple RAID", MustParseGUID("52414944-0000-11AA-AA11-00306543ECAC")},
	{"appleraidoff", "Apple RAID, offline", MustParseGUID("52414944-5F4F-11AA-AA11-00306543ECAC")},
	{"appleboot", "Apple Boot", MustParseGUID("426F6F74-0000-11AA-AA11-00306543ECAC")},
	{"applelabel", "Apple Label", MustParseGUID("4C616265-6C00-11AA-AA11-00306543ECAC")},
	{"appletv", "Apple TV Recovery", MustParseGUID("5265636F-7665-11AA-AA11-00306543ECAC")},
	{"applecs", "Apple Core Storage", MustParseGUID("53746F72-6167-11AA-AA11-00306543ECAC")},
	{"applesrs", "Apple SoftRAID Status", MustParseGUID("B6FA30DA-92D2-4A9A-96F1-871EC6486200")},
	{"applesrscr", "Apple SoftRAID Scratch", MustParseGUID("2E313465-19B9-463F-8126-8A7993773801")},
	{"applesrv", "Apple SoftRAID Volume", MustParseGUID("FA709C7E-65B1-4593-BFD5-E71D61DE9B02")},
	{"applesrc", "Apple SoftRAID Cache", MustParseGUID("BBBA6DF5-F46F-4A89-8F59-8765B2727503")},
	{"solarisboot", "Solaris Boot", MustParseGUID("6A82CB45-1DD2-11B2-99A6-080020736631")},
	{"solarisroot", "Solaris Root", MustParseGUID("6A85CF4D-1DD2-11B2-99A6-080020736631")},
	{"solarisswap", "Solaris Swap", MustParseGUID("6A87C46F-1DD2-11B2-99A6-080020736631")},
	{"solarisbakup", "Solaris Backup", MustParseGUID("6A8B642B-1DD2-11B2-99A6-080020736631")},
	{"solarisusr", "Solaris /usr", MustParseGUID("6A898CC3-1DD2-11B2-99A6-080020736631")},
	{"solarisvar", "Solaris /var", MustParseGUID("6A8EF2E9-1DD2-11B2-99A6-080020736631")},
	{"solarishome", "Solaris /home", MustParseGUID("6A90BA39-1DD2-11B2-99A6-080020736631")},
	{"solarisalt", "Solaris Alternate sector", MustParseGUID("6A9283A5-1DD2-11B2-99A6-080020736631")},
	{"solaris", "Solaris Reserved", MustParseGUID("6A945A3B-1DD2-11B2-99A6-080020736631")},
	{"solaris", "Solaris Reserved", MustParseGUID("6A9630D1-1DD2-11B2-99A6-080020736631")},
	{"solaris", "Solaris Reserved", MustParseGUID("6A980767-1DD2-11B2-99A6-080020736631")},
	{"solaris", "Solaris Reserved", MustParseGUID("6A96237F-1DD2-11B2-99A6-080020736631")},
	{"solaris", "Solaris Reserved", MustParseGUID("6A8D2AC7-1DD2-11B2-99A6-080020736631")},
	{"nbsdswap", "NetBSD Swap", MustParseGUID("49F48D32-B10E-11DC-B99B-0019D1879648")},
	{"nbsdffs", "NetBSD FFS", MustParseGUID("49F48D5A-B10E-11DC-B99B-0019D1879648")},
	{"nbsdlfs", "NetBSD LFS", MustParseGUID("49F48D82-B10E-11DC-B99B-0019D1879648")},
	{"nbsdraid", "NetBSD RAID", MustParseGUID("49F48DAA-B10E-11DC-B99B-0019D1879648")},
	{"nbsdcat", "NetBSD Concatenated", MustParseGUID("2DB519C4-B10F-11DC-B99B-0019D1879648")},
	{"nbsdcrypt", "NetBSD Encrypted", MustParseGUID("2DB519EC-B10F-11DC-B99B-0019D1879648")},
	{"chromeoskern", "ChromeOS kernel", MustParseGUID("FE3A2A5D-4F32-41A7-B725-ACCC3285A309")},
	{"chromeosroot", "ChromeOS rootfs", MustParseGUID("3CB8E202-3B7E-47DD-8A3C-7FF2A13CFCEC")},
	{"chromeos", "ChromeOS future use", MustParseGUID("2E0A753D-9E48-43B0-8337-B15192CB1B5E")},
	{"haikubfs", "Haiku BFS", MustParseGUID("42465331-3BA3-10F1-802A-4861696B7521")},
	{"midbsdboot", "MidnightBSD Boot", MustParseGUID("85D5E45E-237C-11E1-B4B3-E89A8F7FC3A7")},
	{"midbsddata", "MidnightBSD Data", MustParseGUID("85D5E45A-237C-11E1-B4B3-E89A8F7FC3A7")},
	{"midbsdswap", "MidnightBSD Swap", MustParseGUID("85D5E45B-237C-11E1-B4B3-E89A8F7FC3A7")},
	{"midbsdufs", "MidnightBSD Unix File System", MustParseGUID("0394EF8B-237E-11E1-B4B3-E89A8F7FC3A7")},
	{"midbsdvvm", "MidnightBSD Vinum volume manager", MustParseGUID("85D5E45C-237C-11E1-B4B3-E89A8F7FC3A7")},
	{"midbsdzfs", "MidnightBSD ZFS", MustParseGUID("85D5E45D-237C-11E1-B4B3-E89A8F7FC3A7")},
	{"cephjournal", "Ceph Journal", MustParseGUID("45B0969E-9B03-4F30-B4C6-B4B80CEFF106")},
	{"cephcrypt", "Ceph dm-crypt Encrypted Journal", MustParseGUID("45B0969E-9B03-4F30-B4C6-5EC00CEFF106")},
	{"cephosd", "Ceph OSD", MustParseGUID("4FBD7E29-9D25-41B8-AFD0-062C0CEFF05D")},
	{"cephdsk", "Ceph disk in creation", MustParseGUID("89C57F98-2FE5-4DC0-89C1-F3AD0CEFF2BE")},
	{"cephcryptosd", "Ceph dm-crypt OSD", MustParseGUID("89C57F98-2FE5-4DC0-89C1-5EC00CEFF2BE")},
	{"openbsd", "OpenBSD Data", MustParseGUID("824CC7A0-36A8-11E3-890A-952519AD3F61")},
	{"qnx6", "QNX6 Power-safe file system", MustParseGUID("CEF5A9AD-73BC-4601-89F3-CDEEEEE321A1")},
	{"plan9", "Plan 9", MustParseGUID("C91818F9-8025-47AF-89D2-F030D7000C2C")},
	{"vmwarecore", "vmkcore (coredump partition)", MustParseGUID("9D275380-40AD-11DB-BF97-000C2911D1B8")},
	{"vmwarevmfs", "VMFS filesystem partition", MustParseGUID("AA31E02A-400F-11DB-9590-000C2911D1B8")},
	{"vmwarersrv", "VMware Reserved", MustParseGUID("9198EFFC-31C0-11DB-8F78-000C2911D1B8")},
	{"androidiabootldr", "Android-IA bootloader", MustParseGUID("2568845D-2332-4675-BC39-8FA5A4748D15")},
	{"androidiabootldr2", "Android-IA bootloader 2", MustParseGUID("114EAFFE-1552-4022-B26E-9B053604CF84")},
	{"androidiaboot", "Android-IA boot", MustParseGUID("49A4D17F-93A3-45C1-A0DE-F50B2EBE2599")},
	{"androidiarecovery", "Android-IA recovery", MustParseGUID("4177C722-9E92-4AAB-8644-43502BFD5506")},
	{"androidiamisc", "Android-IA misc", MustParseGUID("EF32A33B-A409-486C-9141-9FFB711F6266")},
	{"androidiametadata", "Android-IA metadata", MustParseGUID("20AC26BE-20B7-11E3-84C5-6CFDB94711E9")},
	{"androidiasystem", "Android-IA system", MustParseGUID("38F428E6-D326-425D-9140-6E0EA133647C")},
	{"androidiacache", "Android-IA cache", MustParseGUID("A893EF21-E428-470A-9E55-0668FD91A2D9")},
	{"androidiadata", "Android-IA data", MustParseGUID("DC76DDA9-5AC1-491C-AF42-A82591580C0D")},
	{"androidiapersistent", "Android-IA persistent", MustParseGUID("EBC597D0-2053-4B15-8B64-E0AAC75F4DB1")},
	{"androidiafactory", "Android-IA factory", MustParseGUID("8F68CC74-C5E5-48DA-BE91-A0C8C15E9C80")},
	{"androidiafastboot", "Android-IA fastboot", MustParseGUID("767941D0-2085-11E3-AD3B-6CFDB94711E9")},
	{"androidiaoem", "Android-IA OEM", MustParseGUID("AC6D7924-EB71-4DF8-B48D-E267B27148FF")},
	{"onieboot", "Onie Boot", MustParseGUID("7412F7D5-A156-4B13-81DC-867174929325")},
	{"oniecfg", "Onie Config", MustParseGUID("D4E6E2CD-4469-46F3-B5CB-1BFF57AFC149")},
	{"ppcboot", "Prep boot", MustParseGUID("9E1A2D38-C612-4316-AA26-8B49521E5A8B")},
	{"fdesktopboot", "Extended Boot Partition", MustParseGUID("BC13C2FF-59E6-4262-A352-B275FD6F7172")},
}
