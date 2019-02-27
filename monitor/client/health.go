package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/sdeoras/jwt"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func checkHealth(services ...string) (map[string]string, error) {
	jwtRequestor := jwt.NewRequestor(os.Getenv("JWT_SECRET_KEY"))
	out := make(map[string]string)

	for _, service := range services {
		request := new(healthpb.HealthCheckRequest)
		request.Service = service

		b, err := proto.Marshal(request)
		if err != nil {
			return nil, err
		}

		req, err := jwtRequestor.Request(http.MethodPost,
			"https://"+os.Getenv("GOOGLE_GCF_DOMAIN")+
				"/"+ProjectName+"/health", nil, b)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%s:%s. Mesg:%s",
				"expected status 200 OK, got", resp.Status, string(b))
		}

		resp.Body.Close()

		response := new(healthpb.HealthCheckResponse)
		if err := proto.Unmarshal(b, response); err != nil {
			return nil, err
		}

		out[service] = response.Status.String()
	}

	return out, nil
}
