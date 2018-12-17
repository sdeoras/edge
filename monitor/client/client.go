package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sdeoras/home-automation/grpc/monitor"
	"google.golang.org/grpc"
)

//go:generate ./build.sh

const (
	PORT  = "50051"
	CHUNK = 0x100000
)

func main() {
	server := flag.String("host", "", "Host IP address")
	flag.Parse()

	if *server == "" {
		log.Fatal("Please enter server IP using --host")
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*server+":"+PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	c := monitor.NewMonitorClient(conn)

	stream, err := c.Query(context.Background(), &monitor.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)
	tag := "monitor.jpg"
	once := new(sync.Once)

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		once.Do(func() {
			tag = filepath.Join("/tmp", data.Tag, tag)
			_ = os.MkdirAll(filepath.Join("/tmp", data.Tag), 0755)
		})

		if nn, err := bw.Write(data.Data); err != nil {
			log.Fatal(err)
		} else {
			if nn != len(data.Data) {
				log.Fatal("not all data written")
			}
		}
	}

	if err := bw.Flush(); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(tag, bb.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
