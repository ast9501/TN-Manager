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
sudo ./TN-Manager
```
You can visit swagger doc on:
```
http://<server-ip>:8080/swagger/index.html
```