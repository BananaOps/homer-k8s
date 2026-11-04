package ec2

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"

	homerconfig "github.com/BananaOps/homer-k8s/pkg/config"
)

// fakeEC2 serves pre-built pages keyed by NextToken ("" for the first page).
type fakeEC2 struct {
	pages map[string]*ec2.DescribeInstancesOutput
	err   error
	calls atomic.Int32
	input *ec2.DescribeInstancesInput
}

func (f *fakeEC2) DescribeInstances(_ context.Context, in *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	f.calls.Add(1)
	f.input = in
	if f.err != nil {
		return nil, f.err
	}
	return f.pages[aws.ToString(in.NextToken)], nil
}

func instance(id, ip, name string) types.Instance {
	i := types.Instance{InstanceId: aws.String(id)}
	if ip != "" {
		i.PrivateIpAddress = aws.String(ip)
	}
	if name != "" {
		i.Tags = []types.Tag{{Key: aws.String("Env"), Value: aws.String("prod")}, {Key: aws.String("Name"), Value: aws.String(name)}}
	}
	return i
}

func TestListRunningInstances_FollowsPagination(t *testing.T) {
	client := &fakeEC2{pages: map[string]*ec2.DescribeInstancesOutput{
		"": {
			Reservations: []types.Reservation{{Instances: []types.Instance{instance("i-1", "10.0.0.1", "web")}}},
			NextToken:    aws.String("page2"),
		},
		"page2": {
			Reservations: []types.Reservation{{Instances: []types.Instance{instance("i-2", "", "")}}},
		},
	}}

	got, err := ListRunningInstances(context.Background(), client, "eu-west-1")
	if err != nil {
		t.Fatal(err)
	}
	if client.calls.Load() != 2 {
		t.Fatalf("expected 2 DescribeInstances calls, got %d", client.calls.Load())
	}
	want := []Instance{
		{Name: "web", PrivateIPAddress: "10.0.0.1", InstanceID: "i-1", Region: "eu-west-1"},
		{Name: "i-2", PrivateIPAddress: "", InstanceID: "i-2", Region: "eu-west-1"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d instances, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("instance %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
	f := client.input.Filters
	if len(f) != 1 || aws.ToString(f[0].Name) != "instance-state-name" || f[0].Values[0] != "running" {
		t.Errorf("unexpected filters: %+v", f)
	}
}

func TestListRunningInstances_ReturnsError(t *testing.T) {
	client := &fakeEC2{err: errors.New("access denied")}
	_, err := ListRunningInstances(context.Background(), client, "eu-west-1")
	if err == nil || !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestBuildGroups(t *testing.T) {
	groups := BuildGroups([]Instance{
		{Name: "web", PrivateIPAddress: "10.0.0.1", InstanceID: "i-1", Region: "eu-west-1"},
		{Name: "noip", InstanceID: "i-2", Region: "us-east-1"},
	})
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	web := groups[0]
	if web.Name != "i-1" || len(web.Items) != 2 {
		t.Fatalf("unexpected group: %+v", web)
	}
	if web.Items[0].Url != "http://10.0.0.1" {
		t.Errorf("unexpected ip url %q", web.Items[0].Url)
	}
	wantConsole := "https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#InstanceDetails:instanceId=i-1"
	if web.Items[1].Url != wantConsole {
		t.Errorf("console url = %q, want %q", web.Items[1].Url, wantConsole)
	}

	// An instance without private IP only gets the console link.
	if len(groups[1].Items) != 1 || groups[1].Items[0].Icon != "fa-brands fa-aws" {
		t.Errorf("expected only the console link, got %+v", groups[1].Items)
	}
}

func TestWritePage(t *testing.T) {
	dir := t.TempDir()
	groups := BuildGroups([]Instance{{Name: "web", PrivateIPAddress: "10.0.0.1", InstanceID: "i-1", Region: "eu-west-1"}})

	changed, err := WritePage(dir, "ec2", groups)
	if err != nil || !changed {
		t.Fatalf("first write: changed=%v err=%v", changed, err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "ec2.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var page homerconfig.HomerPageConfig
	if err := yaml.Unmarshal(raw, &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Services) != 1 || page.Services[0].Items[0].SubTitle != "10.0.0.1" {
		t.Errorf("unexpected page content:\n%s", raw)
	}

	changed, err = WritePage(dir, "ec2", groups)
	if err != nil || changed {
		t.Fatalf("second identical write: changed=%v err=%v", changed, err)
	}
}

func TestWritePage_RejectsInvalidNames(t *testing.T) {
	for _, name := range []string{"", "../escape", "a/b", `a\b`, "page.yml"} {
		if _, err := WritePage(t.TempDir(), name, nil); err == nil {
			t.Errorf("expected error for page name %q", name)
		}
	}
}

func TestWritePage_MissingDir(t *testing.T) {
	if _, err := WritePage(filepath.Join(t.TempDir(), "missing"), "ec2", nil); err == nil {
		t.Fatal("expected error for missing config dir")
	}
}

func TestSyncer_Sync(t *testing.T) {
	dir := t.TempDir()
	s := &Syncer{
		Client: &fakeEC2{pages: map[string]*ec2.DescribeInstancesOutput{
			"": {Reservations: []types.Reservation{{Instances: []types.Instance{instance("i-1", "10.0.0.1", "web")}}}},
		}},
		Region:    "eu-west-1",
		ConfigDir: dir,
		PageName:  "servers",
		Log:       logr.Discard(),
	}
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "servers.yml")); err != nil {
		t.Fatalf("page file not written: %v", err)
	}

	s.Client = &fakeEC2{err: errors.New("throttled")}
	if err := s.Sync(context.Background()); err == nil {
		t.Fatal("expected Sync to surface the AWS error")
	}
}

func TestSyncer_RunStopsOnCancelAndSurvivesErrors(t *testing.T) {
	client := &fakeEC2{err: errors.New("boom")}
	s := &Syncer{Client: client, ConfigDir: t.TempDir(), PageName: "ec2", Log: logr.Discard()}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx, time.Millisecond) }()

	deadline := time.After(2 * time.Second)
	for client.calls.Load() < 3 {
		select {
		case <-deadline:
			t.Fatal("Run did not keep looping after errors")
		case <-time.After(time.Millisecond):
		}
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("HOMER_TEST_VAR", "")
	if got := envOrDefault("HOMER_TEST_VAR", "def"); got != "def" {
		t.Errorf("got %q", got)
	}
	t.Setenv("HOMER_TEST_VAR", "set")
	if got := envOrDefault("HOMER_TEST_VAR", "def"); got != "set" {
		t.Errorf("got %q", got)
	}
}
