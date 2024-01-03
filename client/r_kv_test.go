package client_test

import (
	"context"
	"fmt"
	"github.com/RealFax/RedQueen/client"
	"github.com/RealFax/RedQueen/internal/hack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestKvClient_Set(t *testing.T) {
	c, err := client.New(context.Background(), []string{
		"127.0.0.1:5230",
	}, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	for i := 0; i < 100000; i++ {
		sErr := c.Set(context.Background(), hack.String2Bytes(fmt.Sprintf("KEY-%d", i)), []byte{}, 120, nil)
		if sErr != nil {
			t.Error(sErr)
		}
		t.Logf("Task: %d done", i)
	}

	t.Log("ok")
}
