package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os/exec"

	"github.com/sdeoras/api/pb"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

//go:generate ./build.sh

const (
	ADDR = "127.0.0.1"
	PORT = "50051"
)

type server struct {
	tag string
}

func (s *server) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	response := new(healthpb.HealthCheckResponse)
	response.Status = healthpb.HealthCheckResponse_SERVING
	return response, nil
}

func (s *server) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return nil
}

func (s *server) Query(_ *pb.Empty, stream pb.Monitor_QueryServer) error {
	img, err := getImage()
	if err != nil {
		return err
	}

	start := 0
	stop := 4096
	for {
		stop = start + 4096

		if start >= len(img) {
			break
		}

		if stop > len(img) {
			stop = len(img)
		}

		data := new(pb.Data)
		data.Data = img[start:stop]
		data.Tag = s.tag
		if err := stream.Send(data); err != nil {
			return err
		}

		start += len(data.Data)
	}
	return nil
}

func getImage() ([]byte, error) {
	b, err := exec.Command("raspistill",
		"-w", "299",
		"-h", "299",
		"-t", "10",
		"-q", "50",
		"-o", "-",
	).Output()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func main() {
	tag := flag.String("tag", "rpi-zero-default", "tag string identifier")
	addr := flag.String("host", ADDR, "address for this service")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	srv := &server{tag: *tag}
	pb.RegisterMonitorServer(s, srv)
	healthpb.RegisterHealthServer(s, srv)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
