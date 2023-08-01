package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/ast9501/TN-Manager/docs" // include swagger doc
)

var sysLogger *log.Logger

// Map bridgeName to vxlanInterface
var BridgeMap map[string]string = make(map[string]string)

// @title Bridge API
// @version 1.0
// @description API endpoints for managing bridges and interfaces.
// @BasePath /
func main() {
	sysLogger = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)

	router := gin.Default()

	// register swagger url
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		v1.GET("/bridge", getBridge)
		v1.POST("/bridge/:bridge_name", addBridge)
		v1.POST("/interface", addInterface)
		v1.POST("/vxlan/:bridge_name", addVxlanBridge)
		v1.DELETE("/vxlan/:bridge_name", delVxlanBridge)
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

	// Setup vxlan interface
	sysLogger.Println("Create VXLAN interface: ", request.VxlanInterface)
	cmd := exec.Command("ip", "link", "add", request.VxlanInterface, "type", "vxlan", "id", request.VxlanId, "remote", request.RemoteIp, "dev", request.BindInterface)
	err := cmd.Run()
	if err != nil {
		sysLogger.Println("Failed to create vxlan interface: ", err)
		c.String(http.StatusInternalServerError, "Failed to create vxlan interface")
		return
	}

	sysLogger.Println("Set VXLAN mtu to 1400")
	cmd = exec.Command("ip", "link", "set", "dev", request.VxlanInterface, "mtu", "1400")
	err = cmd.Run()
	if err != nil {
		sysLogger.Println("Failed to set vxlan mtu: ", err)
		c.String(http.StatusInternalServerError, "Failed to set vxlan mtu")
		return
	}

	// Create bridge
	sysLogger.Println("Create bridge: ", vxlanBridgeName)
	cmd = exec.Command("brctl", "addbr", vxlanBridgeName)
	err = cmd.Run()
	if err != nil {
		sysLogger.Println("Failed to create bridge: ", err)
		c.String(http.StatusInternalServerError, "Failed to create bridge")
		return
	}

	// Add vxlan interface to bridge
	sysLogger.Println("Add VXLAN interface to bridge")
	cmd = exec.Command("brctl", "addif", vxlanBridgeName, request.VxlanInterface)
	err = cmd.Run()
	if err != nil {
		sysLogger.Println("Failed to add VXLAN interface to bridge: ", err)
		c.String(http.StatusInternalServerError, "Failed to add VXLAN interface to bridge")
		return
	}

	// Configure bridge ip
	sysLogger.Println("Set bridge ip: ", request.LocalBridgeIp)
	cmd = exec.Command("ip", "addr", "add", request.LocalBridgeIp, "dev", vxlanBridgeName)
	err = cmd.Run()
	if err != nil {
		sysLogger.Println("Failed to set bridge ip: ", err)
		c.String(http.StatusInternalServerError, "Failed to set bridge ip")
		return
	}

	cmd = exec.Command("ip", "link", "set", request.VxlanInterface, "up")
	err = cmd.Run()
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

	BridgeMap[vxlanBridgeName] = request.VxlanInterface

	response := fmt.Sprintf("Bridge %s created successfully", vxlanBridgeName)
	c.String(http.StatusCreated, response)
}

//TODO: GetVxlanBridge

//TODO: DeleteVxlanBridge
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
		cmd := exec.Command("ip", "link", "set", vxlanIf, "down")
		err := cmd.Run()
		if err != nil {
			sysLogger.Println("Failed to disable device ", vxlanIf)
			c.String(http.StatusInternalServerError, "Failed to disable vxlan interface")
			return
		}

		cmd = exec.Command("ip", "link", "set", vxlanBridgeName, "down")
		err = cmd.Run()
		if err != nil {
			sysLogger.Println("Failed to disable device ", vxlanBridgeName)
			c.String(http.StatusInternalServerError, "Failed to disable bridge")
			return
		}

		// Remove vxlan interface
		cmd = exec.Command("ip", "link", "del", vxlanIf)
		err = cmd.Run()
		if err != nil {
			sysLogger.Println("Failed to delete device ", vxlanIf)
			c.String(http.StatusInternalServerError, "Failed to delete device")
			return
		}

		// Remove bridge
		cmd = exec.Command("brctl", "delbr", vxlanBridgeName)
		err = cmd.Run()
		if err != nil {
			sysLogger.Println("Failed to remove bridge ", vxlanBridgeName)
			c.String(http.StatusInternalServerError, "Failed to remove bridge")
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

	err := createBridge(bridgeName)
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

// createBridge creates a Linux bridge with the given name.
func createBridge(bridgeName string) error {
	cmd := exec.Command("brctl", "addbr", bridgeName)
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("ip", "link", "set", "dev", bridgeName, "up")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
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
