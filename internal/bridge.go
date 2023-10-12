package internal

import (
	"log"

	"github.com/vishvananda/netlink"
)

var internalLogger *log.Logger

func GetBridge(bridgeName string) (netlink.Link, error) {
	bridgeLink, err := netlink.LinkByName(bridgeName)
	if err != nil {
		// bridge doesn't exist
		return nil, err
	}

	return bridgeLink, err
}

func CreateBridge(bridgeName string) (*netlink.Bridge, error) {
	bridgeLink := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: bridgeName,
		},
	}

	err := netlink.LinkAdd(bridgeLink)
	if err != nil {
		internalLogger.Println("Failed to create bridge:", err)
		return nil, err
	}

	return bridgeLink, nil
}

func SetBridgeIp(ipv4Addr string, bridgeLink netlink.Link) error {
	addr, err := netlink.ParseAddr(ipv4Addr)
	if err != nil {
		internalLogger.Panic("Failed to parse ipv4 addr, ", err)
		return err
	}
	err = netlink.AddrAdd(bridgeLink, addr)
	if err != nil {
		internalLogger.Println("Failed to add ipv4 to bridge:", err)
		return err
	}
	return nil
}

func DelBridge(bridgeName string) error {
	bridgeLink, err := netlink.LinkByName(bridgeName)
	if err != nil {
		internalLogger.Println("Failed to get bridge interface: ", err)
		return err
	}

	err = netlink.LinkDel(bridgeLink)
	if err != nil {
		internalLogger.Println("Failed to delete bridge interface: ", err)
		return err
	}

	return nil
}
