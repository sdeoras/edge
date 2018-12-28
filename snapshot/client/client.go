package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"

	"github.com/sdeoras/edge/grpc/snapshot"
	"google.golang.org/grpc"
)

//go:generate ./build.sh

const (
	PORT  = "50051"
	CHUNK = 0x100000
)

func getImage() ([]byte, error) {
	b, err := exec.Command("raspistill", "-t", "2s", "-o", "-").Output()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func main() {
	server := flag.String("host", "", "Host IP address")
	tag := flag.String("tag", "generic", "Image tag")
	flag.Parse()

	if *server == "" {
		log.Fatal("Please enter server IP using --host")
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*server+":"+PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// capture image
	b, err := getImage()
	if err != nil {
		log.Fatal(err)
	}

	c := snapshot.NewSnapshotClient(conn)

	stream, err := c.Send(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	n := 0
	for {
		if len(b[n:]) < CHUNK {
			if err := stream.Send(&snapshot.Data{Data: b[n:], Tag: *tag}); err != nil {
				log.Fatal(err)
			}
			break
		} else {
			if err := stream.Send(&snapshot.Data{Data: b[n : n+CHUNK], Tag: *tag}); err != nil {
				log.Fatal(err)
			}
			n += CHUNK
		}
	}

	recv, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Written", recv.N, "bytes over GPRC and then to GCP")
}
