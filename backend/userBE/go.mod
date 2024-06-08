module userBE

go 1.22.0

replace clcum/protobuf => ../../protobuf

require (
	clcum/protobuf v0.0.0-00010101000000-000000000000
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	google.golang.org/grpc v1.64.0
)

require (
	github.com/felixge/httpsnoop v1.0.3 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)
