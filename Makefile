all: generate-api-doc build-app

SOURCE = .
INSTDIR = bin

generate-api-doc:
	swag init -g $(SOURCE)/main.go

build-app:
	go build -o $(INSTDIR)/TN-Manager $(SOURCE)/main.go

build-image:
	sudo docker build -t alan0415/tn-manager:v0.2 .

clean:
	rm -rf bin
	rm -rf docs