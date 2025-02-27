package katnplib

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"unsafe"

	"github.com/google/uuid"
	"github.com/vishvananda/netlink"
	"github.com/containernetworking/plugins/pkg/ns"
	"golang.org/x/sys/unix"
)

const (
	tapPrefix	= "tap"
	tapLen		= 8
)

/* Struct that holds the ethtool link settings, converted from C */
type ethtoolCmd struct {
	Cmd           uint32
	Supported     uint32
	Advertising   uint32
	Speed         uint16
	Duplex        uint8
	Port          uint8
	PhyAddress    uint8
	Transceiver   uint8
	Autoneg       uint8
	MdioSupport   uint8
	Maxtxpkt      uint32
	Maxrxpkt      uint32
	SpeedHi       uint16
	EthTpMdix     uint8
	EthTpMdixCtrl uint8
	LpAdvertising uint32
	Reserved      [2]uint32
}

func randomTapName() string {
	randomUuid, _ := uuid.NewRandom()

	return tapPrefix + strings.Replace(randomUuid.String(), "-", "", -1)[:tapLen]
}

/* Set the tap interface speed through ioctl, using the ethtool endpoints */
func SetTapSpeed(tapIface string) (error) {
	fd, err := syscall.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	ecmd := &ethtoolCmd {
		Cmd: 0x00000002, 	/* ETHTOOL_SSET - Set ethtool_link_settings */	
		Duplex: 0x01,		/* FULL_DUPLEX */
	}

	/* 10G - Expressed as 1Mb */
	fullSpeed := uint32(10000)
	ecmd.Speed = uint16(fullSpeed & 0xffff)
	ecmd.SpeedHi = uint16(fullSpeed >> 16)

	ifreq := &netlink.Ifreq{Data: uintptr(unsafe.Pointer(ecmd))}
	copy(ifreq.Name[:unix.IFNAMSIZ-1], tapIface)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), netlink.SIOCETHTOOL, uintptr(unsafe.Pointer(ifreq)))
	if errno != 0 {
		return fmt.Errorf("errno=%v", errno)
	}

	return nil
}

func CreateTap(macAddress net.HardwareAddr) (string, int, error) {
	tapName := randomTapName()
	
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = tapName
	linkAttrs.HardwareAddr = macAddress

	if err := netlink.LinkAdd(&netlink.Tuntap{
		LinkAttrs: linkAttrs,
		Flags: netlink.TUNTAP_NO_PI,
		Mode: netlink.TUNTAP_MODE_TAP,
	}); err != nil {
		return "", -1, err
	}

	iface, err := netlink.LinkByName(tapName)
	if err != nil {
		return "", -1, fmt.Errorf("failed to lookup %q: %v", tapName, err)
	}

	err = SetTapSpeed(tapName)
	if err != nil {
		return "", -1, fmt.Errorf("failed to set tap speed %q: %v", tapName, err)
	}

	return tapName, iface.Attrs().Index, nil
}

func DeleteTap(tapIface string, tapIfaceIdx int, nsPath string) error {
	netns, err := ns.GetNS(nsPath)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", nsPath, err)
	}
	defer netns.Close()

	err = netns.Do(func(hostNS ns.NetNS) error {
		iface, err := netlink.LinkByIndex(tapIfaceIdx)
		if err != nil {
			return fmt.Errorf("failed to lookup %q in %q: %v", tapIface, hostNS.Path(), err)
		}

		if err := netlink.LinkDel(iface); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
