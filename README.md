# iFundRobot

This is a microservice using gRPC protocal to operate supervisor services.

# Requisites
  * GO 1.16 above
  * protobuf 3.15 above
  * gRPC 


## Generate gRPC code
 
```bash
protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
protos/ifundrobot.proto
```

# Build Project

## Build client and server
  `make all`