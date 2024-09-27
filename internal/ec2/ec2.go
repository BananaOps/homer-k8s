package ec2

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	homerv1alpha1 "github.com/BananaOps/homer-k8s/api/v1alpha1"
)

// Define logger
var logger logr.Logger

type Instance struct {
	Name             string
	PrivateIpAddress string
	InstanceId       string
	Region           string
}

type HomerServices struct {
	Services []homerv1alpha1.Group `json:"services,omitempty"`
}

func discoverEC2Instances() []homerv1alpha1.Group {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := ec2.NewFromConfig(cfg)

	filters := []types.Filter{
		{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		},
	}

	// Define parameters to describe instances
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	// Request DescribeInstances
	result, err := svc.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to describe instances, %v", err)
	}

	var listInstances []Instance

	// for each instance in the result, get the name and the private ip address
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			// Get the name of the instance
			title := instance.InstanceId
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					title = tag.Value
					break
				}
			}

			listInstances = append(listInstances, Instance{
				Name:             *title,
				PrivateIpAddress: *instance.PrivateIpAddress,
				InstanceId:       *instance.InstanceId,
				Region:           svc.Options().Region,
			})
		}
	}

	var homerGroups = make([]homerv1alpha1.Group, 0, len(listInstances))

	for _, instance := range listInstances {
		homerGroups = append(homerGroups, homerv1alpha1.Group{
			Name: instance.InstanceId,
			Items: []homerv1alpha1.Item{
				{
					Name:     instance.Name,
					Icon:     "fa-solid fa-server",
					Url:      fmt.Sprintf("http://%s", instance.PrivateIpAddress),
					SubTitle: instance.PrivateIpAddress,
					Target:   "_blank",
				},
				{
					Name:     instance.Name,
					Icon:     "fa-brands fa-aws",
					Url:      fmt.Sprintf("https://%s.console.aws.amazon.com/ec2/home?region=%s#InstanceDetails:instanceId=%s", instance.Region, instance.Region, instance.InstanceId),
					SubTitle: instance.PrivateIpAddress,
					Target:   "_blank",
				},
			},
		},
		)
	}

	return homerGroups
}

func HomerEc2() {

	var HomerServices HomerServices
	ec2group := discoverEC2Instances()

	HomerServices.Services = ec2group

	d, _ := yaml.Marshal(HomerServices)

	pageName := os.Getenv("HOMER_EC2_PAGE_NAME")
	if pageName == "" {
		pageName = "ec2"
	}

	configDir := os.Getenv("HOMER_CONFIG_DIR")
	if configDir == "" {
		configDir = "/assets"
	}

	filePath := fmt.Sprintf("%s/%s.yml", configDir, pageName)
	err := os.WriteFile(filepath.Clean(filePath), d, 0600)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("config ec2 discovery updated")
}

func ControllerEc2() {
	logger.Info("controller ec2 running...")
	// Create a new ticker to run the function every 30 seconds
	interval := time.NewTicker(30 * time.Second)
	for range interval.C {
		HomerEc2()
	}
}

// Init logger slog for json and output to stdout
func init() {
	opts := zap.Options{
		Development: false,
	}
	logger = zap.New(zap.UseFlagOptions(&opts))

}
