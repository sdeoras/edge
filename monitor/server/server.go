package main

import (
	"flag"
	"log"
	"net"
	"os/exec"

	"github.com/sdeoras/home-automation/grpc/monitor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:generate ./build.sh

const (
	ADDR = ":50051"
)

type server struct {
	tag string
}

func (s *server) Query(_ *monitor.Empty, stream monitor.Monitor_QueryServer) error {
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

		data := new(monitor.Data)
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
	b, err := exec.Command("raspistill", "-t", "2s", "-o", "-").Output()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func main() {
	tag := flag.String("tag", "rpi-zero-default", "tag string identifier")
	flag.Parse()

	lis, err := net.Listen("tcp", ADDR)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	monitor.RegisterMonitorServer(s, &server{tag: *tag})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
