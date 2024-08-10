package api

import (
	"context"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
)

func NewLightsailAPI(logger core.ILogger, accessKeyId, secretAccessKey string) core.ServerAPI {
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
	logger core.ILogger
}

func (api *LightsailAPI) GetServers(ctx context.Context) ([]*core.Server, error) {
	output, err := api.client.GetInstances(ctx, &lightsail.GetInstancesInput{})
	if err != nil {
		return nil, err
	}

	servers := []*core.Server{}
	for _, instance := range output.Instances {
		if instance.Name != nil {
			server := core.Server{
				Name: *instance.Name,
			}
			servers = append(servers, &server)
		}
	}
	return servers, nil
}

func (api *LightsailAPI) RebootServer(ctx context.Context, server *core.Server) error {
	output, err := api.client.RebootInstance(ctx, &lightsail.RebootInstanceInput{
		InstanceName: &server.Name,
	})
	api.logger.Infof("reboot server result %+v", output)
	return err
}
