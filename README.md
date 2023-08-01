# TN-Manager
> Simple manager to manage TransportNetwork
Simulate a simple transport network to brige subnetwork.
Create a Transport Network (Linux Bridge), and add subnet interface (Linux Bridge) on demand.

## Build
* Install swaggo
```
go install github.com/swaggo/swag/cmd/swag@v1.16.2
```

* Build(Update) swagger api doc
```
swag init
```

* Build TN-Manager
```
go build -o TN-Manager main.go
```

## Usage
```
# use `8080` as default service port, or you can provide `--port=XXX` to use another port
sudo ./TN-Manager
```

You can visit swagger doc on:
```
http://<server-ip>:<server-port>/swagger/index.html
```

### Manage VXLAN bridge
#### Create new bridge with vxlan interface
This api will setup a new vxlan interface, create a new Linux bridge and bind the vxlan interface to the bridge.

* Sample payload
  * bindInterface: Local interface name to establish vxlan tunnel
  * localBrIp: the ip assign to new bridge (local)
  * remoteIp: Remote interface ip to establish vxlan tunnel
  * vxlanId: the vxlan ID
  * vxlanInterface: the new vxlan interface name 
```
#URL: /api/v1/vxlan/{vxlan_bridge_name}
{
  "bindInterface": "ens3",
  "localBrIp": "192.168.3.222/24",
  "remoteIp": "192.168.101.176",
  "vxlanId": "100",
  "vxlanInterface": "vxlan100"
}
```

#### Delete VXLAN bridge
This api will remove the vxlan bridge (Linux bridge and VXLAN interface)
```
#URL: /api/v1/vxlan/{vxlan_bridge_name}
```