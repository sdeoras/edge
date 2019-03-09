package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"github.com/sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/sdeoras/api"
	"github.com/sdeoras/jwt"
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
	health := flag.Bool("health", false, "show server health and exit")
	download := flag.Bool("download", false, "download image and exit")
	model := flag.String("model", "", "model/version to use")
	expectedOut := flag.String("expect", "", "expected output label")
	skipNotification := flag.Bool("skip-notification", false, "skip user notification")
	flag.Parse()

	modelName, modelVersion := filepath.Split(*model)
	if modelName == "" {
		modelName = modelVersion
		modelVersion = "v1"
	} else {
		modelName = strings.TrimRight(modelName, "/")
	}

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
		out, err := checkHealth("infer", "email")
		if err != nil {
			log.Fatal(err)
		}

		c := healthpb.NewHealthClient(conn)
		response, err := c.Check(context.Background(), &healthpb.HealthCheckRequest{
			Service: "check",
		})

		if err != nil {
			log.Fatal(err)
		}

		out["server"] = response.Status.String()

		jb, err := json.Marshal(out)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jb))

		return
	}

	c := api.NewMonitorClient(conn)

	stream, err := c.Query(context.Background(), &api.Empty{})
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

	if *download {
		if err := ioutil.WriteFile("/tmp/monitor.jpg", bb.Bytes(), 0644); err != nil {
			log.Fatal(err)
		}

		logrus.Info("wrote /tmp/monitor.jpg")
		return
	}

	request := new(api.InferImageRequest)
	request.Images = make([]*api.Image, 1)
	request.Images[0] = new(api.Image)
	request.Images[0].Name = modelName
	request.Images[0].Data = bb.Bytes()
	request.ModelName = modelName
	request.ModelVersion = modelVersion

	b, err := proto.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	jwtRequestor := jwt.NewManager(os.Getenv("JWT_SECRET_KEY"))

	req, err := jwtRequestor.Request(http.MethodPost, "https://"+os.Getenv("GOOGLE_GCF_DOMAIN")+
		"/"+ProjectName+"/"+NameInfer, nil, b)

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

	logrus.WithField("model", filepath.Join(modelName, modelVersion)).
		Info(response.Outputs[0].Label)
	if response.Outputs[0].Label == *expectedOut {
		return
	}

	if *skipNotification {
		logrus.Info("skipping notification")
		return
	}

	sendRequest := &api.EmailRequest{
		FromName:  os.Getenv("EMAIL_FROM_NAME"),
		FromEmail: os.Getenv("EMAIL_FROM_EMAIL"),
		ToName:    os.Getenv("EMAIL_TO_NAME"),
		ToEmail:   os.Getenv("EMAIL_TO_EMAIL"),
		Subject:   fmt.Sprintf("%s result is %s", modelName, response.Outputs[0].Label),
		Body:      []byte(fmt.Sprintf("<strong>%s result is %s</strong>",
			modelName, response.Outputs[0].Label)),
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

	logrus.Info("sent email", sendResponse.StatusCode)
}
