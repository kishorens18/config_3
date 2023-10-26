package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pb "go_config/proto"
	"log"
	"net"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedMyServiceServer
}
type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type RequestData struct {
	ID    int            `json:"_id"`
	Name  string         `json:"name"`
	Value []KeyValuePair `json:"Value"`
}

func (s *server) InsertData(ctx context.Context, req *pb.Request) (*emptypb.Empty, error) {

	var value interface{}
	err := json.Unmarshal(req.Value.Value, &value)
	if err != nil {
		return nil, err
	}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("kishore").Collection("nithish")

	document := bson.M{
		"Name":   req.Name,
		"Config": []interface{}{value},
	}

	result, err := collection.InsertOne(context.Background(), document)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)

	return &emptypb.Empty{}, nil
}

func (s *server) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.GetDataResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("kishore").Collection("nithish")

	filter := bson.M{"_id": id}
	var wholeDocumentResult bson.M
	err = collection.FindOne(context.TODO(), filter).Decode(&wholeDocumentResult)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = collection.FindOne(context.TODO(), filter).Decode(&wholeDocumentResult)
	if err != nil {
		fmt.Println("Error in Whole Document")
		fmt.Println(err)
		return nil, err
	}

	fmt.Println("Whole:")
	fmt.Println(wholeDocumentResult)

	var specificConfigResult bson.M
	configFilter := bson.M{
		"_id": id,
		"Config": bson.M{
			"$elemMatch": bson.M{
				"key": req.Key,
			},
		},
	}
	err = collection.FindOne(context.TODO(), configFilter).Decode(&specificConfigResult)
	if err != nil {
		fmt.Println("Error in Specific")
		fmt.Println(err)
		return nil, err
	}

	// // Retrieve the value for the specified key
	// configArray, ok := specificConfigResult["Config"].(primitive.A)
	// if !ok {
	//     fmt.Println("Config key not found")
	//     return nil, errors.New("config key not found")
	// }

	// var value interface{}
	// for _, configItem := range configArray {
	//     item, ok := configItem.(primitive.M)
	//     if !ok {
	//         continue
	//     }

	//     // Check if the key matches the requested key
	//     if key, keyExists := item["key"].(string); keyExists && key == req.Key {
	//         value = item["value"]
	//         break
	//     }
	// }

	// // Print and return the value
	// fmt.Println("Specific Value:", value)

	var specifiedKeyValuePair map[string]interface{}

	// Retrieve the value for the specified key
	configArray, ok := specificConfigResult["Config"].(primitive.A)
	if !ok {
		fmt.Println("Config key not found")
		return nil, errors.New("config key not found")
	}

	for _, configItem := range configArray {
		item, ok := configItem.(primitive.M)
		if !ok {
			continue
		}

		if key, keyExists := item["key"].(string); keyExists && key == req.Key {
			specifiedKeyValuePair = map[string]interface{}{
				"Config": []map[string]interface{}{
					{
						"key":   item["key"],
						"value": item["value"],
					},
				},
				"Name": specificConfigResult["Name"],
				"_id":  specificConfigResult["_id"].(primitive.ObjectID).Hex(),
			}
			break
		}
	}

	fmt.Println("SpecifiedKeyValuePair:")
	fmt.Println(specifiedKeyValuePair)

	response := &pb.GetDataResponse{}
	response.GDRA = append(response.GDRA, &pb.Application{
		Id:   id.Hex(),
		Name: specificConfigResult["Name"].(string),
	})

	return response, nil

}

func (s *server) AddConfig(ctx context.Context, req *pb.AddConfigRequest) (*emptypb.Empty, error) {
	var value interface{}
	err := json.Unmarshal(req.Value.Value, &value)
	if err != nil {
		return nil, err
	}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer client.Disconnect(context.Background())
	collection := client.Database("kishore").Collection("nithish")
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	filter := bson.M{"_id": id}
	update := bson.M{
		"$push": bson.M{"Config": bson.M{"key": req.Key, "value": value}},
	}

	res, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(res.UpsertedID)

	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":5001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("Listening")
	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterMyServiceServer(s, &server{})
	if err2 := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen: %v", err2)
	}
}
