package api

import (
	"context"
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
)

func NewLightsailAPI(logger core.Logger, accessKeyId, secretAccessKey string) legacy.ServerAPI {
	client := lightsail.New(lightsail.Options{
		Region: "eu-central-1",
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     accessKeyId,
				SecretAccessKey: secretAccessKey,
			}, nil
		}),
	})
	return &LightsailAPI{
		client: client,
		logger: logger,
	}
}

type LightsailAPI struct {
	client *lightsail.Client
	logger core.Logger
}

func (api *LightsailAPI) GetServers(ctx context.Context) ([]*legacy.Server, error) {
	output, err := api.client.GetInstances(ctx, &lightsail.GetInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("get instances request failed: %w", err)
	}

	servers := []*legacy.Server{}
	for _, instance := range output.Instances {
		if instance.Name != nil {
			server := legacy.Server{
				Name: *instance.Name,
			}
			servers = append(servers, &server)
		}
	}
	return servers, nil
}

func (api *LightsailAPI) RebootServer(ctx context.Context, server *legacy.Server) error {
	output, err := api.client.RebootInstance(ctx, &lightsail.RebootInstanceInput{
		InstanceName: &server.Name,
	})
	api.logger.Info("reboot server result", "output", output)
	return err
}
