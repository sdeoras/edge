package main

import (
	"github.com/sdeoras/health"
	"github.com/sdeoras/jwt"
	"net/http"
	"os"
	"path/filepath"
)

func checkHealth(services ...string) (map[string]string, error) {
	jwtManager := jwt.NewManager(os.Getenv("JWT_SECRET_KEY"))
	healthProvider := health.NewProvider(health.OutputProto, jwtManager, nil)

	out := make(map[string]string)

	for _, service := range services {
		req, err := healthProvider.Request(service, "https://"+filepath.Join(
			os.Getenv("GOOGLE_GCF_DOMAIN"),
			ProjectName,
			health.StdRoute))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		_, mesg, err := healthProvider.Response(resp)
		if err != nil {
			return nil, err
		}

		out[service] = mesg
	}

	return out, nil
}
