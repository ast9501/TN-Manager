package internal

import (
	"net"
	"os"
	"strconv"

	"github.com/vishvananda/netlink"
)

func CreateVxlan(vxlanIntfName, vxlanId, localIp, remoteIp string) (*netlink.Vxlan, error) {
	vxlanIdInt, _ := strconv.Atoi(vxlanId)
	vxlanLink := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name: vxlanIntfName,
		},
		VxlanId: vxlanIdInt,
		SrcAddr: net.ParseIP(localIp),
		Group:   net.ParseIP(remoteIp),
	}

	err := netlink.LinkAdd(vxlanLink)
	if err != nil {
		internalLogger.Println("Failed to create VXLAN interface:", err)
		return nil, err
	}

	return vxlanLink, nil
}

func SetVxlanMaster(vxlanLink *netlink.Vxlan, bridgeLink *netlink.Bridge) error {
	err := netlink.LinkSetMaster(vxlanLink, bridgeLink)
	if err != nil {
		internalLogger.Println("Failed to set master:", err)
		return err
	}

	return nil
}

func SetVxlanDown(vxlanIntfName string) error {
	vxlanLink, err := netlink.LinkByName(vxlanIntfName)
	if err != nil {
		internalLogger.Println("Failed to get VXLAN interface:", err)
		return err
	}

	err = netlink.LinkSetDown(vxlanLink)
	if err != nil {
		internalLogger.Println("Failed to bring down VXLAN interface:", err)
		return err
	}

	err = netlink.LinkSetNoMaster(vxlanLink)
	if err != nil {
		println("Failed to remove from master:", err)
		os.Exit(1)
	}

	return nil
}

func DelVxlan(vxlanIntfName string) error {
	vxlanLink, err := netlink.LinkByName(vxlanIntfName)
	if err != nil {
		internalLogger.Println("Failed to get VXLAN interface:", err)
		return err
	}

	err = netlink.LinkDel(vxlanLink)
	if err != nil {
		internalLogger.Println("Failed to delete VXLAN interface:", err)
		return err
	}

	return nil
}
