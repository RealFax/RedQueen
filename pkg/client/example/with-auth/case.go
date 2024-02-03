package main

import (
	"context"
	"github.com/RealFax/RedQueen/pkg/client"
	"github.com/RealFax/RedQueen/pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	// create basic-auth client
	authBroker := grpcutil.NewBasicAuthClient("admin", "123456")

	c, err := client.New(context.Background(), []string{
		"127.0.0.1:3230",
		"127.0.0.1:4230",
		"127.0.0.1:5230",
	},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// setup auth client interceptor
		grpc.WithUnaryInterceptor(authBroker.Unary),
		grpc.WithStreamInterceptor(authBroker.Stream),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if _, err = c.Get(context.Background(), []byte("Key1"), nil); err != nil {
		log.Fatal("client get error:", err)
	}
}
