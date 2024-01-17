package efi

type SignatureData struct {
	Owner GUID
	Data  []byte
}

type SignatureList struct {
	Type       GUID
	ListSize   uint32
	HeaderSize uint32
	Size       uint32
}

const (
	IMAGE_EXECUTION_AUTHENTICATION     = 0x00000007
	IMAGE_EXECUTION_AUTH_UNTESTED      = 0x00000000
	IMAGE_EXECUTION_AUTH_SIG_FAILED    = 0x00000001
	IMAGE_EXECUTION_AUTH_SIG_PASSED    = 0x00000002
	IMAGE_EXECUTION_AUTH_SIG_NOT_FOUND = 0x00000003
	IMAGE_EXECUTION_AUTH_SIG_FOUND     = 0x00000004
	IMAGE_EXECUTION_POLICY_FAILED      = 0x00000005
	IMAGE_EXECUTION_INITIALIZED        = 0x00000008
)

type (
	Md5Hash    [16]uint8
	Sha1Hash   [20]uint8
	Sha224Hash [28]uint8
	Sha256Hash [32]uint8
	Sha384Hash [48]uint8
	Sha512Hash [64]uint8
)
