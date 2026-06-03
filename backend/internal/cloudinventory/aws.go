package cloudinventory

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// HasAWS reports whether AWS credentials are configured.
func HasAWS() bool {
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_ROLE_ARN") != "" || os.Getenv("AWS_REGION") != ""
}

// AWSProvider inventories EC2 and RDS via paginated Describe APIs.
type AWSProvider struct{}

func NewAWSProvider() *AWSProvider { return &AWSProvider{} }

func (p *AWSProvider) Name() string { return "aws" }

func (p *AWSProvider) ListAssets(ctx context.Context, opts ListOptions) ([]Asset, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	now := time.Now().UTC()
	var out []Asset

	ec2Client := ec2.NewFromConfig(cfg)
	var ec2Token *string
	pages := 0
	for {
		if opts.MaxPages > 0 && pages >= opts.MaxPages {
			break
		}
		resp, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			NextToken: ec2Token,
			MaxResults: aws.Int32(int32(clampPage(opts.PageSize))),
		})
		if err != nil {
			return nil, fmt.Errorf("ec2 describe: %w", err)
		}
		for _, res := range resp.Reservations {
			for _, inst := range res.Instances {
				out = append(out, instanceToAsset(inst, now))
			}
		}
		pages++
		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}
		ec2Token = resp.NextToken
	}

	rdsClient := rds.NewFromConfig(cfg)
	var rdsMarker *string
	pages = 0
	for {
		if opts.MaxPages > 0 && pages >= opts.MaxPages {
			break
		}
		resp, err := rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
			Marker: rdsMarker,
			MaxRecords: aws.Int32(int32(clampPage(opts.PageSize))),
		})
		if err != nil {
			return nil, fmt.Errorf("rds describe: %w", err)
		}
		for _, db := range resp.DBInstances {
			out = append(out, rdsToAsset(db, now))
		}
		pages++
		if resp.Marker == nil || *resp.Marker == "" {
			break
		}
		rdsMarker = resp.Marker
	}
	return out, nil
}

func instanceToAsset(inst ec2types.Instance, now time.Time) Asset {
	name := ""
	for _, t := range inst.Tags {
		if aws.ToString(t.Key) == "Name" {
			name = aws.ToString(t.Value)
			break
		}
	}
	if name == "" {
		name = aws.ToString(inst.InstanceId)
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = aws.ToString(inst.Placement.AvailabilityZone)
	}
	return Asset{
		ID: "aws-ec2-" + aws.ToString(inst.InstanceId),
		Provider: "aws", Type: "ec2", Name: name, Region: region,
		Status: string(inst.State.Name), UpdatedAt: now,
		Tags: map[string]string{"source": "aws-sdk"},
	}
}

func rdsToAsset(db rdstypes.DBInstance, now time.Time) Asset {
	return Asset{
		ID: "aws-rds-" + aws.ToString(db.DBInstanceIdentifier),
		Provider: "aws", Type: "rds",
		Name: aws.ToString(db.DBInstanceIdentifier),
		Region: aws.ToString(db.AvailabilityZone),
		Status: aws.ToString(db.DBInstanceStatus),
		UpdatedAt: now,
		Tags: map[string]string{"engine": aws.ToString(db.Engine), "source": "aws-sdk"},
	}
}

func clampPage(n int) int32 {
	if n <= 0 {
		return 100
	}
	if n > 1000 {
		return 1000
	}
	return int32(n)
}
