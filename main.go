package main

import (
	"crypto/rand"
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

// @title Bridge API
// @version 1.0
// @description API endpoints for managing bridges and interfaces.
// @host 192.168.0.189:8080
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
	}

	router.Run(":8080")
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
