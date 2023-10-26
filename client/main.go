package main

import (
	"context"
	"encoding/json"
	pb "go_config/proto"
	"log"

	"google.golang.org/grpc"
	
)

type KeyValuePair struct {
	Key   string `json:"key"`
	Value bool `json:"value"`
}

var client pb.MyServiceClient

func main() {
	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	client = pb.NewMyServiceClient(conn)

	// InsertData()
	// AddConfig()
	GetData()
}

func InsertData() {
	name := "Kyc-App"
	value := KeyValuePair{
		Key:   "1",
		Value: true,
	}
	valueJSON, err := json.Marshal(value)
	if err != nil {
		log.Fatalf("Failed to marshal value to JSON: %v", err)
	}
	req := &pb.Request{
		Name: name,
		Value: string(valueJSON),
	}
	_, err = client.InsertData(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}
}

func GetData() {

	id := "653a2d746d1225d37852be97"
	key := "5"
	req := &pb.GetDataRequest{
		Id:  id,
		Key: key,
	}
	_, err := client.GetData(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to get data: %v", err)
	}
}

func AddConfig() {
	id := "653a2d746d1225d37852be97"
	key := "5"
	value := []string{"1", "2", "3"}
	valueJSON, err := json.Marshal(value)
	if err != nil {
		log.Fatalf("Failed to marshal value to JSON: %v", err)
	}
	req := &pb.AddConfigRequest{
		Id:  id,
		Key: key,
		Value: string(valueJSON),
	}
	_, err = client.AddConfig(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}
}
