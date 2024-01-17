package bdxde

type UPD struct {
	/**Offset 0x0000
	 **/
	Signature uint64
	/**Offset 0x0008
	 **/
	Reserved uint64
	/**Offset 0x0010
	 **/
	UnusedUpdSpace0 [16]uint8
	/**Offset 0x0020
	   Debug serial port resource type. Select 'None' to have FSP generate no output. Select 'I/O' to have FSP generate output via a legacy I/O output (i.e. 0x2f8/0x3f8).
	**/
	SerialPortType uint8
	/**Offset 0x0021
	   16550 compatible serial port resource base address. (I/O or MMIO base address)
	**/
	SerialPortAddress uint32
	/**Offset 0x0025
	   Select 'Yes' to have FSP configure the specified UART.
	**/
	SerialPortConfigure uint8
	/**Offset 0x0026
	   If 'Configure Serial Port' is set to 'Yes', this will be the baud rate that the UART located at 'Serial Port Base' is configured to use. *: Not all ES2 part support this baudrate, check Sighting Report before change it.
	**/
	SerialPortBaudRate uint8
	/**Offset 0x0027
	   Select 'Yes' to have FSP initialize this controller. Note: If 'Yes' is selected, this controller will be mapped to the legacy IO port 0x3f8.
	**/
	SerialPortControllerInit0 uint8
	/**Offset 0x0028
	   Select 'Yes' to have FSP initialize this controller. Note: If 'Yes' is selected, this controller will be mapped to the legacy IO port 0x2f8.
	**/
	SerialPortControllerInit1 uint8
	/**Offset 0x0029
	   Set your bifurcation option for IOU1 port
	**/
	ConfigIOU1_PciPort3 uint8
	/**Offset 0x002A
	   Set your bifurcation option for IOU2 port
	**/
	ConfigIOU2_PciPort1 uint8
	/**Offset 0x002B
	   S0 = System will return to S0 state (boot) after power is re-applied. S5 = System will return to the S5 state (except if it was in S4, in which case it will return to S4). In the S5 state, the only enabled wake event is the Power Button or any enabled wake event that was preserved through the power failure.
	**/
	PowerStateAfterG3 uint8
	/**Offset 0x002C
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort1 uint8
	/**Offset 0x002D
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort2 uint8
	/**Offset 0x002E
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort3 uint8
	/**Offset 0x002F
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort4 uint8
	/**Offset 0x0030
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort5 uint8
	/**Offset 0x0031
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort6 uint8
	/**Offset 0x0032
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort7 uint8
	/**Offset 0x0033
	   Enable/Disable PCH PCie Ports. Port bifurcation options set in Fuse Straps decide the final enabling of PCIe ports
	**/
	PchPciPort8 uint8
	/**Offset 0x0034
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort1 uint8
	/**Offset 0x0035
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort2 uint8
	/**Offset 0x0036
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort3 uint8
	/**Offset 0x0037
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort4 uint8
	/**Offset 0x0038
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort5 uint8
	/**Offset 0x0039
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort6 uint8
	/**Offset 0x003A
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort7 uint8
	/**Offset 0x003B
	   Enable/Disable the HotPlug for PCH PCie Ports
	**/
	HotPlug_PchPciPort8 uint8
	/**Offset 0x003C
	   Enable or disable the EHCI controller at 00.1d.00.
	**/
	Ehci1Enable uint8
	/**Offset 0x003D
	   Enable or disable the EHCI controller at 00.1a.00.
	**/
	Ehci2Enable uint8
	/**Offset 0x003E
	   Enable or disable Intel(r) Hyper-Threading Technology.
	**/
	HyperThreading uint8
	/**Offset 0x003F
	   Set debug print output level.
	**/
	DebugOutputLevel uint8
	/**Offset 0x0040
	   Halt and Lock the TCO Timer.
	**/
	TcoTimerHaltLock uint8
	/**Offset 0x0041
	   Enable/Disable processor Turbo Mode.
	**/
	TurboMode uint8
	/**Offset 0x0042
	   Select the performance state that should be set before OS hand-off.
	**/
	BootPerfMode uint8
	/**Offset 0x0043
	   PCI-Express Root Port ASPM Setting
	**/
	PciePort1aAspm uint8
	/**Offset 0x0044
	   PCI-Express Root Port ASPM Setting
	**/
	PciePort1bAspm uint8
	/**Offset 0x0045
	   PCI-Express Root Port ASPM Setting
	**/
	PciePort3aAspm uint8
	/**Offset 0x0046
	   PCI-Express Root Port ASPM Setting
	**/
	PciePort3bAspm uint8
	/**Offset 0x0047
	   PCI-Express Root Port ASPM Setting
	**/
	PciePort3cAspm uint8
	/**Offset 0x0048
	   PCI-Express Root Port ASPM Setting
	**/
	PciePort3dAspm uint8
	/**Offset 0x0049
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort1Aspm uint8
	/**Offset 0x004A
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort2Aspm uint8
	/**Offset 0x004B
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort3Aspm uint8
	/**Offset 0x004C
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort4Aspm uint8
	/**Offset 0x004D
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort5Aspm uint8
	/**Offset 0x004E
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort6Aspm uint8
	/**Offset 0x004F
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort7Aspm uint8
	/**Offset 0x0050
	   PCH PCIe Root Port ASPM Setting
	**/
	PchPciePort8Aspm uint8
	/**Offset 0x0051
	   Enable this option to allow DFX Lock Bits to remain clear.
	**/
	DFXEnable uint8
	/**Offset 0x0052
	   Enable/Disable the PCH Thermal Device (D31:F6).
	**/
	ThermalDeviceEnable uint8
	/**Offset 0x0053
	 **/
	UnusedUpdSpace1 [88]uint8
	/**Offset 0x00AB
	   Enable/disable DDR ECC Support.
	**/
	MemEccSupport uint8
	/**Offset 0x00AC
	   Select the memory type supported by this platform.
	**/
	MemDdrMemoryType uint8
	/**Offset 0x00AD
	   Force the Rank Multiplication factor for LRDIMM.
	**/
	MemRankMultiplication uint8
	/**Offset 0x00AE
	   Run the Rank Margin Tool after memory training.
	**/
	MemRankMarginTool uint8
	/**Offset 0x00AF
	   Enable data scrambling.
	**/
	MemScrambling uint8
	/**Offset 0x00B0
	   Self refresh mode.
	**/
	MemRefreshMode uint8
	/**Offset 0x00B1
	   Select MC ODT Mode.
	**/
	MemMcOdtOverride uint8
	/**Offset 0x00B2
	   Enable/Disable DDR4 Command Address Parity
	**/
	MemCAParity uint8
	/**Offset 0x00B3
	   Configure Thermal Throttling Mode.
	**/
	MemThermalThrottling uint8
	/**Offset 0x00B4
	   Configures CKE and related Memory Power Savings Features
	**/
	MemPowerSavingsMode uint8
	/**Offset 0x00B5
	   Configure Memory Electrical Throttling
	**/
	MemElectricalThrottling uint8
	/**Offset 0x00B6
	   Select Page Policy
	**/
	MemPagePolicy uint8
	/**Offset 0x00B7
	   Splits the 0-4GB address space between two sockets, so that both sockets get a chunk of local memory below 4GB.
	**/
	MemSocketInterleaveBelow4G uint8
	/**Offset 0x00B8
	   Select Channel Interleaving setting.
	**/
	MemChannelInterleave uint8
	/**Offset 0x00B9
	   Select Rank Interleaving setting.
	**/
	MemRankInterleave uint8
	/**Offset 0x00BA
	   Select 'Yes' if memory is down. If set to 'Yes', at least one of the following SPD data pointers must also be provided.
	**/
	MemDownEnable uint8
	/**Offset 0x00BB
	   If 'Memory Down' is 'Yes', this is the pointer to the SPD data for Channel 0, DIMM 0. Otherwise, this field is ignored. Specify 0x00000000 if this DIMM is not present.
	**/
	MemDownCh0Dimm0SpdPtr uint32
	/**Offset 0x00BF
	   If 'Memory Down' is 'Yes', this is the pointer to the SPD data for Channel 0, DIMM 1. Otherwise, this field is ignored. Specify 0x00000000 if this DIMM is not present.
	**/
	MemDownCh0Dimm1SpdPtr uint32
	/**Offset 0x00C3
	   If 'Memory Down' is 'Yes', this is the pointer to the SPD data for Channel 1, DIMM 0. Otherwise, this field is ignored. Specify 0x00000000 if this DIMM is not present.
	**/
	MemDownCh1Dimm0SpdPtr uint32
	/**Offset 0x00C7
	   If 'Memory Down' is 'Yes', this is the pointer to the SPD data for Channel 1, DIMM 1. Otherwise, this field is ignored. Specify 0x00000000 if this DIMM is not present.
	**/
	MemDownCh1Dimm1SpdPtr uint32
	/**Offset 0x00CB
	   Select 'Yes' to enable Fast Boot.
	**/
	MemFastBoot uint8
	/**Offset 0x00CC
	   Configure how reads and writes of F0000h-FFFFFh handled.
	**/
	Pam0_hienable uint8
	/**Offset 0x00CD
	   Configure how reads and writes of C0000h-C3FFFh handled.
	**/
	Pam1_loenable uint8
	/**Offset 0x00CE
	   Configure how reads and writes of C4000h-C7FFFh handled.
	**/
	Pam1_hienable uint8
	/**Offset 0x00CF
	   Configure how reads and writes of C8000h-CBFFFh handled.
	**/
	Pam2_loenable uint8
	/**Offset 0x00D0
	   Configure how reads and writes of CC000h-CFFFFh handled.
	**/
	Pam2_hienable uint8
	/**Offset 0x00D1
	   Configure how reads and writes of D0000h-D3FFFh handled.
	**/
	Pam3_loenable uint8
	/**Offset 0x00D2
	   Configure how reads and writes of D4000h-D7FFFh handled.
	**/
	Pam3_hienable uint8
	/**Offset 0x00D3
	   Configure how reads and writes of D8000h-DBFFFh handled.
	**/
	Pam4_loenable uint8
	/**Offset 0x00D4
	   Configure how reads and writes of DC000h-DFFFFh handled.
	**/
	Pam4_hienable uint8
	/**Offset 0x00D5
	   Configure how reads and writes of E0000h-E3FFFh handled.
	**/
	Pam5_loenable uint8
	/**Offset 0x00D6
	   Configure how reads and writes of E4000h-E7FFFh handled.
	**/
	Pam5_hienable uint8
	/**Offset 0x00D7
	   Configure how reads and writes of E8000h-EBFFFh handled.
	**/
	Pam6_loenable uint8
	/**Offset 0x00D8
	   Configure how reads and writes of EC000h-EFFFFh handled.
	**/
	Pam6_hienable uint8
	/**Offset 0x00D9
	   Asynchronous DRAM Refresh - Set to 'Enabled' if DIMMs are battery-backed or if the platform implements saving to some other non-volatile storage medium. Set to 'Enabled (NVDIMMs)' if NVDIMMs will be used on the platform.
	**/
	MemAdr uint8
	/**Offset 0x00DA
	 **/
	MemAdrResumePath uint8
	/**Offset 0x00DB
	   Block all PCIe/South Complex Traffic on ADR. (For V0 and later steppings only).
	**/
	MemBlockScTrafficOnAdr uint8
	/**Offset 0x00DC
	   Specify the i/o port which should be written to when CKE/DDR Reset clamps should be released. Specify '0' to skip.
	**/
	MemPlatformReleaseAdrClampsPort uint16
	/**Offset 0x00DE
	   Specify the value that should be ANDed with the value read from the clamp release i/o port.
	**/
	MemPlatformReleaseAdrClampsAnd uint32
	/**Offset 0x00E2
	   Specify the value that should be ORed with the value read from the clamp release i/o port.
	**/
	MemPlatformReleaseAdrClampsOr uint32
	/**Offset 0x00E6
	 **/
	UnusedUpdSpace2 [24]uint8
	/**Offset 0x00FE
	 **/
	PcdRegionTerminator uint16
}

type VPD struct {
	/**Offset 0x0000
	 **/
	PcdVpdRegionSign uint64
	/**Offset 0x0008
	 **/
	PcdImageRevision uint32
	/**Offset 0x000C
	 **/
	PcdUpdRegionOffset uint32
	/**Offset 0x0010
	 **/
	UnusedVpdSpace0 [16]uint8
	/**Offset 0x0020
	 **/
	PcdFspReservedMemoryLength uint32
}

const (
	FSP_IMAGE_ID  = 0x5F45442D5844425F
	FSP_IMAGE_REV = 0x00000301
)