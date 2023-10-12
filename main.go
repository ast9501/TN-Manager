package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/vishvananda/netlink"

	_ "github.com/ast9501/TN-Manager/docs" // include swagger doc
	"github.com/ast9501/TN-Manager/internal"
)

var sysLogger *log.Logger

// Map bridgeName to vxlanInterface
var BridgeMap map[string]string = make(map[string]string)
var SliceMap map[string]string = make(map[string]string)

// @title Bridge API
// @version 1.0
// @description API endpoints for managing bridges and interfaces.
// @BasePath /
func main() {
	sysLogger = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)

	router := gin.Default()

	// register swagger url
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//TODO: Decouple api functions to another module
	v1 := router.Group("/api/v1")
	{
		v1.GET("/bridge", getBridge)
		v1.POST("/bridge/:bridge_name", addBridge)
		v1.POST("/interface", addInterface)
		v1.POST("/vxlan/:bridge_name", addVxlanBridge)
		v1.POST("/vxlan/:bridge_name/activate", activateVxlanBridge)
		v1.DELETE("/vxlan/:bridge_name", delVxlanBridge)
		v1.POST("/slice/:bridge_name", addSlice)
		v1.DELETE("/slice/:bridge_name", delSlice)
	}

	port := flag.String("port", "8080", "service port")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage：\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "flags：\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	router.Run(":" + *port)
}

// addSlice handles the POST /api/v1/slice/:bridge_name endpoint.
// It add slice (tc rule) to vxlan interface.
//
// @Summary Add slice on interface
// @Description
// @Tags slice
// @Accept json
// @Produce json
// @Param bridge_name path string true "Bridge name"
// @Param request body SliceRequest true "Slice request"
// @Success 204 {string} string "Slice Installed"
// @Router /api/v1/slice/{bridge_name} [post]
func addSlice(c *gin.Context) {
	bridgeName := c.Param("bridge_name")

	var request SliceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}

	sysLogger.Println("bridgeName, ", bridgeName)
	vxlanInterface, ok := BridgeMap[bridgeName]

	if ok {
		sysLogger.Println("Add slice on interface, ", vxlanInterface)
		classId, err := internal.AddQdisc(BridgeMap[bridgeName], request.FlowRate)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to add qdisc, ", err)
			return
		}

		err = internal.AddFilter(vxlanInterface, request.DstIp, strconv.Itoa(int(classId)))
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to add filter, ", err)
			return
		}
	} else {
		sysLogger.Println("Failed to find interface")
		c.String(http.StatusInternalServerError, "Failed to find interface")
		return
	}

	sysLogger.Println("Install slice successful, ", "SliceSD", request.SliceSd, "FlowRate (KB/Sec)", request.FlowRate)
	c.String(http.StatusAccepted, "Install Slice successful")
}

// delSlice handles the DELETE /api/v1/slice/:bridge_name/:sliceSd endpoint.
// It del slice (tc rule) on vxlan interface.
//
// @Summary Del slice on interface
// @Description
// @Tags slice
// @Accept json
// @Produce json
// @Param bridge_name path string true "Bridge name"
// @Param slice_sd path string true "Slice SD identifier"
// @Success 204 {string} string "Slice deletion successful"
// @Router /api/v1/slice/{bridge_name}/{slice_sd}} [delete]
func delSlice(c *gin.Context) {
	sliceSd := c.Param("sliceSd")
	bridge_name := c.Param("bridgeName")

	sysLogger.Println("Delete slice ", "Slice SD", sliceSd, "Bridge", bridge_name)
	c.String(http.StatusAccepted, "Slice deleted")
}

// addVxlanBridge handles the POST /api/v1/vxlan/:bridge_name endpoint.
// It adds a new bridge with vxlan interface.
//
// @Summary Add a new bridge
// @Description Add a new bridge with the given name
// @Tags vxlan
// @Accept json
// @Produce json
// @Param bridge_name path string true "Bridge name"
// @Param request body VxlanInterfaceRequest true "Vxlan Interface request"
// @Success 201 {string} string "Bridge created successfully"
// @Failure 400 {string} string "Invalid bridge name"
// @Router /api/v1/vxlan/{bridge_name} [post]
func addVxlanBridge(c *gin.Context) {
	vxlanBridgeName := c.Param("bridge_name")

	var request VxlanInterfaceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}

	//FIXME: Check if vxlan interface exist, if exist return error

	// Setup vxlan interface
	sysLogger.Println("Create VXLAN interface: ", request.VxlanInterface)

	vxlanLink, err := internal.CreateVxlan(request.VxlanInterface, request.VxlanId, request.BindInterface, request.RemoteIp)

	if err != nil {
		sysLogger.Println("Failed to create vxlan interface: ", err)
		c.String(http.StatusInternalServerError, "Failed to create vxlan interface")
		return
	}

	// Check if bridge exist
	bridgeLink, _ := internal.GetBridge(vxlanBridgeName)
	if bridgeLink == nil {
		bridgeLink, err = internal.CreateBridge(vxlanBridgeName)
		if err != nil {
			sysLogger.Println("Failed to create bridge: ", err)
			c.String(http.StatusInternalServerError, "Failed to create bridge")
			return
		}
	}

	bridge, isBridge := bridgeLink.(*netlink.Bridge)
	if !isBridge {
		sysLogger.Println("Failed to assert netlink.Bridge")
		c.String(http.StatusInternalServerError, "The specified bridge is not of type netlink.Bridge")
		return
	}

	err = internal.SetVxlanMaster(vxlanLink, bridge)
	if err != nil {
		sysLogger.Println("Failed to bind vxlan interface to bridge: ", err)
		c.String(http.StatusInternalServerError, "Failed to bind vxlan to bridge")
		return
	}

	err = internal.SetBridgeIp(request.LocalBridgeIp, bridgeLink)
	if err != nil {
		sysLogger.Println("Failed to configure bridge ipv4 addr: ", err)
		c.String(http.StatusInternalServerError, "Failed to set bridge ip")
		return
	}

	// Activate bridge and vxlan
	err = netlink.LinkSetUp(vxlanLink)
	if err != nil {
		sysLogger.Println("Failed to activate vxlan interface: ", err)
		c.String(http.StatusInternalServerError, "Failed to enable vxlan interface")
		return
	}

	err = netlink.LinkSetUp(bridgeLink)
	if err != nil {
		sysLogger.Println("Failed to activate bridge: ", err)
		c.String(http.StatusInternalServerError, "Failed to enable bridge")
		return
	}

	BridgeMap[vxlanBridgeName] = request.VxlanInterface

	response := fmt.Sprintf("Bridge %s created successfully", vxlanBridgeName)
	c.String(http.StatusCreated, response)
}

// activateVxlanBridge handles the POST /api/v1/vxlan/:bridge_name/activate endpoint.
// [Deprecated] It activate bridge with vxlan interface.
//
// @Summary [Deprecated] Activate vxlan bridge
// @Description
// @Tags vxlan
// @Accept json
// @Produce json
// @Param bridge_name path string true "Bridge name"
// @Success 204 {string} string "Bridge Activated"
// @Failure 400 {string} string "Invalid bridge name"
// @Router /api/v1/vxlan/{bridge_name}/activate [post]
func activateVxlanBridge(c *gin.Context) {
	vxlanBridgeName := c.Param("bridge_name")

	cmd := exec.Command("ip", "link", "set", BridgeMap[vxlanBridgeName], "up")
	err := cmd.Run()
	if err != nil {
		sysLogger.Println("Failed to enable vxlan interface: ", err)
		c.String(http.StatusInternalServerError, "Failed to enable vxlan interface")
		return
	}

	cmd = exec.Command("ip", "link", "set", vxlanBridgeName, "up")
	err = cmd.Run()
	if err != nil {
		sysLogger.Println("Failed to enable bridge: ", err)
		c.String(http.StatusInternalServerError, "Failed to enable bridge")
		return
	}

	c.String(http.StatusAccepted, "Bridge Activated")
}

//TODO: GetVxlanBridge

// delVxlanBridge handles the DELETE /api/v1/vxlan/:bridge_name endpoint.
// It delete a exist bridge with vxlan interface.
//
// @Summary Delete bridge
// @Description Delete a exist bridge with the given name
// @Tags vxlan
// @Accept json
// @Produce json
// @Param bridge_name path string true "Bridge name"
// @Success 200 {string} string "Bridge delete successfully"
// @Router /api/v1/vxlan/{bridge_name} [delete]
func delVxlanBridge(c *gin.Context) {
	vxlanBridgeName := c.Param("bridge_name")

	if vxlanIf, exist := BridgeMap[vxlanBridgeName]; exist {
		// Disable device
		sysLogger.Println("Disable device")

		// Set vxlan interface down and unbound vxlan from bridge
		err := internal.SetVxlanDown(vxlanIf)
		if err != nil {
			sysLogger.Println("Failed to disable device ", vxlanIf)
			c.String(http.StatusInternalServerError, "Failed to disable vxlan interface")
			return
		}

		// Remove bridge
		err = internal.DelBridge(vxlanBridgeName)
		if err != nil {
			sysLogger.Println("Failed to disable device ", vxlanBridgeName)
			c.String(http.StatusInternalServerError, "Failed to disable bridge")
			return
		}

		// Remove vxlan interface
		err = internal.DelVxlan(vxlanIf)
		if err != nil {
			sysLogger.Println("Failed to delete device ", vxlanIf)
			c.String(http.StatusInternalServerError, "Failed to delete device")
			return
		}

		// Remove from map
		delete(BridgeMap, vxlanBridgeName)
	} else {
		// Bridge not exist
		c.String(http.StatusNotFound, "Bridge not found")
		return
	}

	c.JSON(http.StatusOK, "Success")
}

// getBridge handles the GET /api/v1/bridge endpoint.
// It returns the current bridge and connected interface.
//
// @Summary Get current bridge and connected interface
// @Description Get the current bridge and its connected interface
// @Tags bridge
// @Produce json
// @Success 200 {object} BridgeResponse
// @Router /api/v1/bridge [get]
func getBridge(c *gin.Context) {
	//TODO: retrieve current bridge and interfaces
	bridgeName := "bridge0"
	interfaceName := "eth0"
	response := BridgeResponse{
		Bridge:    bridgeName,
		Interface: interfaceName,
	}
	c.JSON(http.StatusOK, response)
}

// addBridge handles the POST /api/v1/bridge/:bridge_name endpoint.
// It adds a new bridge with the given name.
//
// @Summary Add a new bridge
// @Description Add a new bridge with the given name
// @Tags bridge
// @Accept json
// @Produce json
// @Param bridge_name path string true "Bridge name"
// @Success 200 {string} string "Bridge created successfully"
// @Failure 400 {string} string "Invalid bridge name"
// @Router /api/v1/bridge/{bridge_name} [post]
func addBridge(c *gin.Context) {

	bridgeName := c.Param("bridge_name")

	//err := createBridge(bridgeName)
	_, err := internal.CreateBridge(bridgeName)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Failed to create bridge: %s", err.Error()))
		return
	}

	response := fmt.Sprintf("Bridge %s created successfully", bridgeName)
	c.String(http.StatusOK, response)
}

// addInterface handles the POST /api/v1/interface endpoint.
// It adds a new interface between two bridges.
//
// @Summary Add a new interface
// @Description Add a new interface between two bridges
// @Tags interface
// @Accept json
// @Produce json
// @Param request body InterfaceRequest true "Interface request"
// @Success 200 {string} string "Interface added successfully"
// @Failure 400 {string} string "Invalid request body"
// @Router /api/v1/interface [post]
func addInterface(c *gin.Context) {
	var request InterfaceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}

	// create veth-pair between two Linux bridge
	err := createVethPair(request.Bridge1, request.Bridge2)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Failed to create veth pair: %s", err.Error()))
		return
	}

	response := "Interface added successfully"
	c.String(http.StatusOK, response)
}

// createVethPair creates a veth pair between two Linux bridges.
func createVethPair(bridge1, bridge2 string) error {
	vethName1 := bridge1 + "-veth" + generateRandomString(4)
	vethName2 := bridge2 + "-veth"
	sysLogger.Println("Create veth-pair: ", vethName1, vethName2)

	cmd := exec.Command("ip", "link", "add", "name", vethName1, "type", "veth", "peer", "name", vethName2)
	err := cmd.Run()
	if err != nil {
		sysLogger.Println("Fail to exec create veth-pair command: ", err)
		return err
	}

	cmd = exec.Command("brctl", "addif", bridge1, vethName1)
	err = cmd.Run()
	if err != nil {
		sysLogger.Println("Fail to exec add veth to bridge command: ", err)
		return err
	}

	cmd = exec.Command("brctl", "addif", bridge2, vethName2)
	err = cmd.Run()
	if err != nil {
		sysLogger.Println("Fail to exec add veth to bridge command: ", err)
		return err
	}

	cmd = exec.Command("ip", "link", "set", "dev", vethName1, "up")
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("ip", "link", "set", "dev", vethName2, "up")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	for i := 0; i < length; i++ {
		randomBytes[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	return string(randomBytes)
}

// BridgeResponse represents the response for the getBridge endpoint.
type BridgeResponse struct {
	Bridge    string `json:"bridge"`
	Interface string `json:"interface"`
}

// InterfaceRequest represents the request body for the addInterface endpoint.
type InterfaceRequest struct {
	Bridge1 string `json:"bridge1"`
	Bridge2 string `json:"bridge2"`
}

type VxlanInterfaceRequest struct {
	BindInterface  string `json:"bindInterface"`
	VxlanInterface string `json:"vxlanInterface"`
	VxlanId        string `json:"vxlanId"`
	//LocalBridgeName		string	`json:"localBrName"`
	RemoteIp      string `json:"remoteIp"`
	LocalBridgeIp string `json:"localBrIp"`
}

type SliceRequest struct {
	FlowRate int    `json:"FlowRate"`
	SliceSd  string `json:"SliceSD,omitempty"`
	DstIp    string `json:"DstIP"`
	SrcIp    string `json:"SrcIP"`
}
