package main

import (
	"io"
	"log"
	"net"

	"github.com/sdeoras/home-automation/grpc/inception"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:generate ./build.sh

const (
	ADDR = ":50052"
)

type server struct{}

func (s *server) Query(stream inception.Inception_QueryServer) error {
	buffer := make([]byte, 0, 0)
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		buffer = append(buffer, data.Data...)
	}

	ack, err := tfExec(buffer)
	if err != nil {
		return err
	}
	return stream.SendAndClose(ack)
}

func main() {
	lis, err := net.Listen("tcp", ADDR)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	inception.RegisterInceptionServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
