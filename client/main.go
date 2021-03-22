package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "ifundrobot/protos"
)

const (
	defaultName    = "DEAN.LIN"
	defaultCurreny = "fUSD"
)

func main() {
	iFundServer := os.Getenv("IFUND_SERVER")
	log.Printf("ifund server %s \n", iFundServer)

	// Set up a connection to the server.
	conn, err := grpc.Dial(iFundServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewIfundrobotClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.CreateRobot(ctx, &pb.RobotRequest{Name: name, Currency: defaultCurreny})
	if err != nil {
		log.Fatalf("could not create robot: %v", err)
	}
	log.Printf("CreateRobot: code: %d, message: %s", r.Code, r.Message)

	s, err := c.RobotStatus(ctx, &pb.RobotRequest{Name: name, Currency: defaultCurreny})
	if err != nil {
		log.Fatalf("could not query robot status: %v", err)
	}
	log.Printf("CreateRobot: %d", s.Code)
}
