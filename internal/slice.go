package internal

import (
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	//"github.com/florianl/go-tc/class"
	//"github.com/florianl/go-tc/filter"
)

var qdiscIndex uint16 = 1

func CreateRootQdisc(vxlanLink *netlink.Link) error {

	qdiscAttr := netlink.QdiscAttrs{
		LinkIndex: (*vxlanLink).Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	}

	// Create root qdisc
	rootQdisc := netlink.NewHtb(qdiscAttr)

	if err := netlink.QdiscAdd(rootQdisc); err != nil {
		internalLogger.Panicln("Failed to create root qdisc:", err)
		return err
	}

	return nil
}

func AddQdisc(vxlanName string, flowRate int) (uint16, error) {
	vxlanLink, err := netlink.LinkByName(vxlanName)
	if err != nil {
		internalLogger.Panicln("Failed to get vxlan link, ", err)
		return 0, err
	}

	//TODO: if root qdisc not exist, create it
	if qdiscIndex == 1 {
		err = CreateRootQdisc(&vxlanLink)
		if err != nil {
			internalLogger.Panic("Failed to create vxlan root qdisc, ", err)
			return 0, err
		}
	}

	// Create class
	classAttr := &netlink.ClassAttrs{
		LinkIndex: vxlanLink.Attrs().Index,
		Handle:    netlink.MakeHandle(1, qdiscIndex),
		Parent:    netlink.HANDLE_ROOT, //tc.HandleRoot,
	}

	htbClassAttr := &netlink.HtbClassAttrs{
		Rate: uint64(flowRate * 8 * 1000),
		Ceil: uint64(flowRate * 8 * 1000),
	}

	class := netlink.NewHtbClass(*classAttr, *htbClassAttr)

	if err := netlink.ClassAdd(class); err != nil {
		internalLogger.Panicln("Failed to create class: ", err)
		return 0, err
	}

	qdiscIndex += 1
	return qdiscIndex - 1, nil
}

// Add filter
func AddFilter(vxlanName, dstIP, classId string) error {

	//TODO: Call library to create tc filter
	cmd := exec.Command("tc", "filter", "add", "dev", vxlanName, "parent", "1:", "protocol", "ip", "prio", "1", "u32", "match", "ip", "dst", dstIP, "flowid", "1:"+classId)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		internalLogger.Panicln("Failed to create tc filter, ", err)
		return err
	}

	/*
		filter := &netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: vxlanIndex(vxlanName),
				Parent:    1, // parent 1:
				Protocol:  unix.ETH_P_IP,
				//Prio:      1,
			},
			ClassId: netlink.MakeHandle(1, 32), // same as ClassAttrs.Handle
			Sel: &netlink.TcU32Sel{
				Nkeys: 1,
				Keys: []netlink.TcU32Key{
					{
						Mask: 0xffffffff,                     // fit all bits
						Val:  iPToUint32(net.ParseIP(dstIP)), // set value
						Off:  16,                             // offset
					},
				},
			},
		}

		err := netlink.FilterAdd(filter)

		if err != nil {
			fmt.Println("Failed to add filter, ", err)
			//FIXME: get err = invalid argument
			os.Exit(1)
		}
	*/

	return nil
}

// TODO: del filter

// TODO: del qdisc

func iPToUint32(ipAddr net.IP) uint32 {
	bits := strings.Split(ipAddr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum uint32
	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)

	return sum
}

// Get vxlan interface index
func vxlanIndex(vxlanName string) int {
	link, err := netlink.LinkByName(vxlanName)
	if err != nil {
		println("Failed to get VXLAN interface:", err)
		os.Exit(1)
	}
	return link.Attrs().Index
}
