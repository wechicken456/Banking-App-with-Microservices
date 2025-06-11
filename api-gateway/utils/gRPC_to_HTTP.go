package utils

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteGRPCErrorToHTTP converts a gRPC error to an appropriate HTTP response
// and writes it to the http.ResponseWriter using the predefined error messages
func WriteGRPCErrorToHTTP(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		// If it's not a gRPC status error, return a generic server error
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
		return
	}

	var httpStatus int
	errorMessage := st.Message()

	switch st.Code() {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
		// Handle specific invalid JWT errors
		if errorMessage == "invalid JWT" {
			httpStatus = http.StatusUnauthorized
		}
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict

	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
		// Both not authorized and not authenticated map to 401
	case codes.Internal:
		httpStatus = http.StatusInternalServerError
	default:
		// For any other gRPC code, return internal server error
		httpStatus = http.StatusInternalServerError
		errorMessage = "Internal server error"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(ErrorResponse{Error: errorMessage})
}
