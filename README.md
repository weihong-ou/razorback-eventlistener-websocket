# Razorback Sawtooth Event Listener with WebSocket Server

This repository is meant to stimulate ideas on how to subscribe to sawtooth event listener and show events on UI through websocket connection.

## System Requirements
1. OS Packages
    ```
    sudo apt-get update
    sudo apt-get -y upgrade
    sudo apt install -y zip curl python3 python3-pip pkg-config
    ```

2. Install protobuf compilers (make sure to get a 3.x.x version from [here](https://github.com/protocolbuffers/protobuf/releases))
    ```
    curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip
    unzip protoc-3.7.1-linux-x86_64.zip -d protoc3
    sudo mv protoc3/bin/* /usr/local/bin/
    sudo mv protoc3/include/* /usr/local/include/
    ```

3. Install python's grpcio-tools library
    ```
    sudo su - 
    python3 -m pip install grpcio-tools
    ```

4. Install golang version 1.12
    ```
    wget https://dl.google.com/go/go1.12.2.linux-amd64.tar.gz

    sudo tar -xvf go1.12.2.linux-amd64.tar.gz
    sudo mv go /usr/bin
    ```

    Add the following to your ~/.profile
    ```
    export GOROOT="/usr/bin/go"
    export GOPATH="$HOME/go"

    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
    ```

5. Install go dependencies
    ```
    go install github.com/golang/mock/mockgen && \
    go get -u google.golang.org/grpc \
        github.com/golang/protobuf/protoc-gen-go \
        github.com/satori/go.uuid \
        github.com/pebbe/zmq4 \
        github.com/golang/mock/gomock \
        github.com/hyperledger/sawtooth-sdk-go \
        github.com/jessevdk/go-flags \
        github.com/stretchr/testify/mock \
        github.com/btcsuite/btcd/btcec \
        gopkg.in/yaml.v2

    cd $GOPATH/src/github.com/hyperledger/sawtooth-sdk-go && \
        go generate

    mkdir -p $GOPATH/src/protobuf/events_pb2

    cp $GOPATH/src/github.com/hyperledger/sawtooth-sdk-go/protobuf/events_pb2/* $GOPATH/src/protobuf/events_pb2/

    ```

## Running the subscriber client

1. Calculate the TP prefix

2. To run, start the validator then type the following on the command line:<br>
	```go run events_subcribe_client.go ```

Note: If you're using docker-compose file default IP is already set.<br>
Otherwise, please set global environment variable as
VALIDATOR_URL="tcp://<VALIDATOR-IP>:4004"

Now, the server is listening on <b>ws://localhost:3811</b>