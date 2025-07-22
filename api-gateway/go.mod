module api-gateway

go 1.24.4

require (
	buf.build/gen/go/banking-app/auth/grpc/go v1.5.1-20250618202807-b09a22ade332.2
	buf.build/gen/go/banking-app/auth/protocolbuffers/go v1.36.6-20250618202807-b09a22ade332.1
	github.com/go-chi/chi/v5 v5.2.1
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.73.0
)

require (
	buf.build/gen/go/banking-app/account/grpc/go v1.5.1-20250722020557-5c16a6b391de.2 // indirect
	buf.build/gen/go/banking-app/account/protocolbuffers/go v1.36.1-20250722020557-5c16a6b391de.1 // indirect
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250613105001-9f2d3c737feb.1 // indirect
	github.com/go-chi/cors v1.2.1 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
