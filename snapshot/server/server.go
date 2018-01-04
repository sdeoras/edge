package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sdeoras/home-automation/grpc/snapshot"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:generate ./build.sh

const (
	ADDR = ":50051"
)

var bucket string

type server struct{}

func (s *server) Send(stream snapshot.Snapshot_SendServer) error {
	buffer := make([]byte, 0, 0)
	var tag string
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		buffer = append(buffer, data.Data...)
		tag = data.Tag
	}

	t := time.Now()
	timeStamp := fmt.Sprintf("%d/%d-%02d/%d-%02d-%02d/%s.jpg",
		t.Year(),
		t.Year(), t.Month(),
		t.Year(), t.Month(), t.Day(),
		t.Format(time.RFC3339))
	fileName := tag + "/" + timeStamp
	if n, err := writeToGS(stream.Context(), bucket, fileName, buffer); err != nil {
		return err
	} else {
		return stream.SendAndClose(&snapshot.Ack{N: int64(n)})
	}
}

func writeToGS(ctx context.Context, bucketName, fileName string, buffer []byte) (int, error) {
	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return 0, err
	}

	// Creates a Bucket instance.
	bucket := client.Bucket(bucketName)

	obj := bucket.Object(fileName)
	w := obj.NewWriter(ctx)
	defer w.Close()
	return w.Write(buffer)
}

func main() {
	bkt := flag.String("bucket", "", "Google storage bucket name")
	flag.Parse()
	if *bkt == "" {
		log.Fatal("Please enter a bucket name using --bucket")
	}
	bucket = *bkt
	lis, err := net.Listen("tcp", ADDR)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	snapshot.RegisterSnapshotServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
