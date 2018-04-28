package pei

import "github.com/qeedquan/disktools/efi"

type Services interface {
	Hdr() efi.TableHeader
	InstallPpi()
	ReInstallPpi()
	LocatePpi()
	NotifyPpi()
	GetBootMode()
	SetBootMode()
	GetHobList()
	CreateHobList()
	CreateHob()
	FfsFindNextVolume()
	FfsFindNextFile()
	FfsFindSectionData()
	InstallPeiMemory()
	AllocatePages()
	AllocatePool()
	CopyMem()
	ResetSystem()
	CpuIo()
	PcfCfg()
}
