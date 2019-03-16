package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/sdeoras/health"
)

func checkHealth(services ...string) (map[string]string, error) {
	healthProvider := health.NewProvider(health.OutputProto)

	out := make(map[string]string)

	for _, service := range services {
		req, err := healthProvider.NewHTTPRequest(service, "https://"+filepath.Join(
			os.Getenv("GOOGLE_GCF_DOMAIN"),
			ProjectName,
			health.StdRoute))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		_, mesg, err := healthProvider.ReadResponseAndClose(resp)
		if err != nil {
			return nil, err
		}

		out[service] = mesg
	}

	return out, nil
}
