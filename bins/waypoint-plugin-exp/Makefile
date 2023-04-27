PLUGIN_NAME=exp

ifndef _ARCH
_ARCH := $(shell ./print_arch)
export _ARCH
endif

.PHONY: all

all: protos build

# Generate the Go code from Protocol Buffer definitions
protos:
	@echo ""
	@echo "Build Protos"

	# skipping because protoc doesn't seem to work. will replace with protoc-gen-go anyway
	# protoc -I . --go_out=plugins=grpc:. --go_opt=paths=source_relative ./builder/output.proto
	# protoc -I . --go_out=plugins=grpc:. --go_opt=paths=source_relative ./registry/output.proto
	# protoc -I thirdparty/proto -I . --go_out=plugins=grpc:. --go_opt=paths=source_relative ./platform/output.proto

# Builds the plugin on your local machine
build:
	@echo ""
	@echo "Compile Plugin"

	# Clear the output
	rm -rf ./bin

	GOOS=linux GOARCH=amd64 go build -o ./bin/linux_amd64/waypoint-plugin-${PLUGIN_NAME} ./main.go 
	GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin_amd64/waypoint-plugin-${PLUGIN_NAME} ./main.go 
	# GOOS=windows GOARCH=amd64 go build -o ./bin/windows_amd64/waypoint-plugin-${PLUGIN_NAME}.exe ./main.go 
	# GOOS=windows GOARCH=386 go build -o ./bin/windows_386/waypoint-plugin-${PLUGIN_NAME}.exe ./main.go 

# Install the plugin locally
install:
	@echo ""
	@echo "Installing Plugin"

	cp ./bin/${_ARCH}_amd64/waypoint-plugin-${PLUGIN_NAME}* ${HOME}/.config/.waypoint/plugins/   

# Zip the built plugin binaries
zip:
	zip -j ./bin/waypoint-plugin-${PLUGIN_NAME}_linux_amd64.zip ./bin/linux_amd64/waypoint-plugin-${PLUGIN_NAME}
	zip -j ./bin/waypoint-plugin-${PLUGIN_NAME}_darwin_amd64.zip ./bin/darwin_amd64/waypoint-plugin-${PLUGIN_NAME}
	# zip -j ./bin/waypoint-plugin-${PLUGIN_NAME}_windows_amd64.zip ./bin/windows_amd64/waypoint-plugin-${PLUGIN_NAME}.exe
	# zip -j ./bin/waypoint-plugin-${PLUGIN_NAME}_windows_386.zip ./bin/windows_386/waypoint-plugin-${PLUGIN_NAME}.exe

# Build the plugin using a Docker container
build-docker:
	rm -rf ./releases
	DOCKER_BUILDKIT=1 docker build --output releases --progress=plain .
