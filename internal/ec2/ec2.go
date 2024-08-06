package ec2

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	homerv1alpha1 "github.com/jplanckeel/homer-k8s/api/v1alpha1"
)

// Define logger
var logger logr.Logger
var logContext []interface{} = []interface{}{"controller", "homerservices", "controllerGroup", "homer.bananaops.io", "controllerKind", "HomerServices"}

var quit = make(chan struct{})

type Instance struct {
	Name             string
	PrivateIpAddress string
	InstanceId       string
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
		{
			Name:   aws.String("tag:costGroup"),
			Values: []string{"recette"},
		},
	}

	// Définir les paramètres de la requête pour décrire les instances
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	// Effectuer la requête pour décrire les instances
	result, err := svc.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to describe instances, %v", err)
	}

	var listInstances []Instance

	// Parcourir les réservations et les instances pour extraire les informations souhaitées
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			// Récupérer le nom de l'instance à partir des tags
			instanceName := "N/A"
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					instanceName = *tag.Value
					break
				}
			}

			listInstances = append(listInstances, Instance{
				Name:             instanceName,
				PrivateIpAddress: *instance.PrivateIpAddress,
				InstanceId:       *instance.InstanceId,
			})
		}
	}

	var homerGroups = make([]homerv1alpha1.Group, 0, len(listInstances))

	for _, instance := range listInstances {
		homerGroups = append(homerGroups, homerv1alpha1.Group{
			Name: instance.InstanceId,
			Items: []homerv1alpha1.Item{
				{
					Name:     instance.PrivateIpAddress,
					Icon:     "fa-solid fa-server",
					Url:      fmt.Sprintf("http://%s", instance.PrivateIpAddress),
					SubTitle: instance.Name,
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
	err := os.WriteFile("/assets/recette.yml", d, 0600)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("config recette updated")
}

func ControllerEc2() {
	logger.Info("controller ec2 running...")
	// Créer un ticker qui déclenche toutes les 30 secondes
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
