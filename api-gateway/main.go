package main

import (
	"api-gateway/handler"
	"api-gateway/initialize"
)

func main() {
	initialize.LoadDotEnv()

	accountClient := handler.NewAccountServiceClient()
	if accountClient == nil {
		panic("Failed to create account service client")
	}

}
