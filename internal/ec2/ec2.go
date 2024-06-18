package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/aws"
)


type Item struct {
	Name string
	PrivateIpAddress string
	InstanceId string
	Tag string `json:"tag,omitempty"`
}


func discoverEC2Instances() {

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

			fmt.Printf("Instance ID: %s, State: %s Private IP: %s %s\n",
				 *instance.InstanceId, 
				  instance.State.Name,
				  *instance.PrivateIpAddress,
				  instanceName,
				)
		}
	}
}

func main() {
	discoverEC2Instances()
}
