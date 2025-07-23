module api-gateway

go 1.24.4

require (
	buf.build/gen/go/banking-app/account/grpc/go v1.5.1-20250723200025-4a71901acaf6.2
	buf.build/gen/go/banking-app/account/protocolbuffers/go v1.36.1-20250723200025-4a71901acaf6.1
	buf.build/gen/go/banking-app/auth/grpc/go v1.5.1-20250723180927-4a955af75edb.2
	buf.build/gen/go/banking-app/auth/protocolbuffers/go v1.36.6-20250723180927-4a955af75edb.1
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	google.golang.org/grpc v1.73.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250717185734-6c6e0d3c608e.1 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
