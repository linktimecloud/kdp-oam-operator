# KDP(Kubernetes Data Platform) OAM API-Server

## Overview

## Quick Start

### Project layout
```
./kdp-oam-apiserver
├── Dockerfile                       # apiserver Dockerfile
├── cmd
│ ├── apiserver
│ │ └── main.go                      # apiserver main function
│ └── options
│     └── options.go                 # apiserver config options
├── docs                             # docs
│ └── openapi                        # apiserver openapi document
│     └── swagger.json               # swagger.json
├── makefiles                        # makefiles set
│ ├── ......
│ ├── build-swagger.mk                 # build-swagger.mk
├── pkg                              # 
│ ├── apiserver                      # apiserver core package
│ │ ├── apis                         # apiserver restful apis definition
│ │ │ └── v1
│ │ │     ├── assembler              # conversion and data exchange between DTO and DO(Domain Object)
│ │ │     ├── dto                    # carrier of data transmission
│ │ │     └── webservice             # restful webservice
│ │ ├── config
│ │ │ └── config.go
│ │ ├── domain                       # core business logic
│ │ │ ├── entity                       # model entity
│ │ │ └── service                      # domain service is a piece of business logic composed of multiple entities
│ │ ├── exception                    # business exception code and message
│ │ ├── infrastructure               # provide general technical infrastructure services for other layers,like database and cache
│ │ ├── server.go                    # apiserver start function
│ │ └── utils                        # common utils

```

### Start the server

1. Install the Go 1.19+.
2. Start the server on local

  ```shell
  # Install all dependencies
  go mod tidy

  # Setting the kube config
  export KUBECONFIG="<Specify your kube config>"
  
  # Generate api server swagger docs(./api/openapi/swagger.json)
  make build-swagger

  # Start the server
  make run-server
  ```

Then, you can open the http://127.0.0.1:8000/apidocs/. 
