package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/sdeoras/edge/grpc/monitor"
	"github.com/sdeoras/jwt"
	"github.com/sdeoras/lambda/api"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

//go:generate ./build.sh

const (
	ADDR        = "127.0.0.1"
	PORT        = "50051"
	ProjectName = "lambda"
	NameInfer   = "infer"
	NameEmail   = "email"
)

func main() {
	server := flag.String("host", ADDR, "host IP address")
	port := flag.String("port", PORT, "port number")
	health := flag.Bool("health", false, "show server health")
	flag.Parse()

	if *server == "" {
		log.Fatal("Please enter server IP using --host")
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*server+":"+*port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	if *health {
		c := healthpb.NewHealthClient(conn)
		res, err := c.Check(context.Background(), &healthpb.HealthCheckRequest{
			Service: "check",
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(res.Status)
		return
	}

	c := monitor.NewMonitorClient(conn)

	stream, err := c.Query(context.Background(), &monitor.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

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

	request := &api.InferImageRequest{
		Data:      bb.Bytes(),
		ModelPath: "garageDoorChecker.pb",
		LabelPath: "garageDoorChecker.txt",
	}

	b, err := proto.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	jwtRequestor := jwt.NewRequestor(os.Getenv("JWT_SECRET_KEY"))

	req, err := jwtRequestor.Request(http.MethodPost, "https://"+os.Getenv("GOOGLE_GCF_DOMAIN")+
		"/"+ProjectName+"/"+NameInfer, nil, b)
	req.Method = http.MethodPost

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s:%s. Mesg:%s", "expected status 200 OK, got", resp.Status, string(b))
	}

	response := new(api.InferImageResponse)
	if err := proto.Unmarshal(b, response); err != nil {
		log.Fatal(err)
	}

	fmt.Println("[" + response.Label + "]")
	if response.Label == "Closed" {
		return
	}

	sendRequest := &api.EmailRequest{
		FromName:  os.Getenv("EMAIL_FROM_NAME"),
		FromEmail: os.Getenv("EMAIL_FROM_EMAIL"),
		ToName:    os.Getenv("EMAIL_TO_NAME"),
		ToEmail:   os.Getenv("EMAIL_TO_EMAIL"),
		Subject:   "garage door is open",
		Body:      []byte("<strong>garage door is open</strong>"),
	}

	b, err = proto.Marshal(sendRequest)
	if err != nil {
		log.Fatal(err)
	}

	req, err = jwtRequestor.Request(http.MethodPost, "https://"+os.Getenv("GOOGLE_GCF_DOMAIN")+
		"/"+ProjectName+"/"+NameEmail, nil, b)
	req.Method = http.MethodPost

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s:%s", "expected status 200 OK, got", resp.Status)
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	sendResponse := new(api.EmailResponse)
	if err := proto.Unmarshal(b, sendResponse); err != nil {
		log.Fatal(err)
	}

	if sendResponse.StatusCode != 202 {
		log.Fatal("sending email failed with status code:", sendResponse.StatusCode)
	}
}
