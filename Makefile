all: generate-api-doc build-app

SOURCE = .
INSTDIR = bin

generate-api-doc:
	swag init -g $(SOURCE)/main.go

build-app:
	go build -o $(INSTDIR)/TN-Manager $(SOURCE)/main.go

clean:
	rm -rf bin
	rm -rf docs