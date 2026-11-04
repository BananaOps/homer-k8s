// Package ec2 discovers running EC2 instances and publishes them as a Homer
// page (<pageName>.yml) in the Homer assets directory.
package ec2

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	ctrl "sigs.k8s.io/controller-runtime"

	homerv1alpha1 "github.com/BananaOps/homer-k8s/api/v1alpha1"
	homerconfig "github.com/BananaOps/homer-k8s/pkg/config"
)

const (
	defaultConfigDir = "/assets"
	defaultPageName  = "ec2"
	// DefaultInterval is the delay between two discoveries.
	DefaultInterval = 30 * time.Second
)

// Instance is the subset of an EC2 instance displayed in Homer.
type Instance struct {
	Name             string
	PrivateIPAddress string
	InstanceID       string
	Region           string
}

// Syncer periodically writes the running EC2 instances into a Homer page file.
type Syncer struct {
	Client    ec2.DescribeInstancesAPIClient
	Region    string
	ConfigDir string
	PageName  string
	Log       logr.Logger
}

// NewFromEnv builds a Syncer from the default AWS credential chain and the
// HOMER_CONFIG_DIR / HOMER_EC2_PAGE_NAME environment variables.
func NewFromEnv(ctx context.Context) (*Syncer, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}
	return &Syncer{
		Client:    ec2.NewFromConfig(cfg),
		Region:    cfg.Region,
		ConfigDir: envOrDefault("HOMER_CONFIG_DIR", defaultConfigDir),
		PageName:  envOrDefault("HOMER_EC2_PAGE_NAME", defaultPageName),
		Log:       ctrl.Log.WithName("ec2"),
	}, nil
}

// Run syncs immediately, then every interval, until ctx is cancelled.
// Sync errors are logged and never stop the loop.
func (s *Syncer) Run(ctx context.Context, interval time.Duration) error {
	s.Log.Info("ec2 discovery running", "interval", interval.String(), "page", s.PageName)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := s.Sync(ctx); err != nil {
			s.Log.Error(err, "ec2 discovery failed")
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// Sync discovers the running instances and writes the page file if it changed.
func (s *Syncer) Sync(ctx context.Context) error {
	instances, err := ListRunningInstances(ctx, s.Client, s.Region)
	if err != nil {
		return err
	}
	changed, err := WritePage(s.ConfigDir, s.PageName, BuildGroups(instances))
	if err != nil {
		return err
	}
	if changed {
		s.Log.Info("ec2 page updated", "instances", len(instances))
	}
	return nil
}

// ListRunningInstances returns every running instance, following pagination.
func ListRunningInstances(ctx context.Context, client ec2.DescribeInstancesAPIClient, region string) ([]Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		}},
	}

	var instances []Instance
	paginator := ec2.NewDescribeInstancesPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				instances = append(instances, toInstance(instance, region))
			}
		}
	}
	return instances, nil
}

func toInstance(i types.Instance, region string) Instance {
	id := aws.ToString(i.InstanceId)
	name := id
	for _, tag := range i.Tags {
		if aws.ToString(tag.Key) == "Name" && aws.ToString(tag.Value) != "" {
			name = aws.ToString(tag.Value)
			break
		}
	}
	return Instance{
		Name:             name,
		PrivateIPAddress: aws.ToString(i.PrivateIpAddress),
		InstanceID:       id,
		Region:           region,
	}
}

// BuildGroups turns instances into Homer groups: one group per instance, with a
// link to its private IP (when it has one) and a link to the AWS console.
func BuildGroups(instances []Instance) []homerv1alpha1.Group {
	groups := make([]homerv1alpha1.Group, 0, len(instances))
	for _, instance := range instances {
		var items []homerv1alpha1.Item
		if instance.PrivateIPAddress != "" {
			items = append(items, homerv1alpha1.Item{
				Name:     instance.Name,
				Icon:     "fa-solid fa-server",
				Url:      "http://" + instance.PrivateIPAddress,
				SubTitle: instance.PrivateIPAddress,
				Target:   "_blank",
			})
		}
		items = append(items, homerv1alpha1.Item{
			Name: instance.Name,
			Icon: "fa-brands fa-aws",
			Url: fmt.Sprintf("https://%s.console.aws.amazon.com/ec2/home?region=%s#InstanceDetails:instanceId=%s",
				instance.Region, instance.Region, instance.InstanceID),
			SubTitle: instance.PrivateIPAddress,
			Target:   "_blank",
		})
		groups = append(groups, homerv1alpha1.Group{Name: instance.InstanceID, Items: items})
	}
	return groups
}

// WritePage writes <pageName>.yml into configDir and reports whether the file
// content changed. File access is scoped to configDir through os.Root (CWE-22).
func WritePage(configDir, pageName string, groups []homerv1alpha1.Group) (bool, error) {
	if pageName == "" || strings.ContainsAny(pageName, "/\\.") {
		return false, fmt.Errorf("invalid page name %q: must not contain path separators or dots", pageName)
	}

	data, err := yaml.Marshal(homerconfig.HomerPageConfig{Services: groups})
	if err != nil {
		return false, err
	}

	root, err := os.OpenRoot(filepath.Clean(configDir))
	if err != nil {
		return false, fmt.Errorf("cannot open config root %q: %w", configDir, err)
	}
	defer func() { _ = root.Close() }()

	fileName := pageName + ".yml"
	if existing, err := root.ReadFile(fileName); err == nil && bytes.Equal(existing, data) {
		return false, nil
	}
	if err := root.WriteFile(fileName, data, 0600); err != nil {
		return false, err
	}
	return true, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
