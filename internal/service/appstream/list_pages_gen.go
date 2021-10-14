// Code generated by "internal/generate/listpages/main.go -ListOps=DescribeFleets,DescribeImageBuilders,DescribeStacks -Export=yes"; DO NOT EDIT.

package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
)

func DescribeFleetsPages(conn *appstream.AppStream, input *appstream.DescribeFleetsInput, fn func(*appstream.DescribeFleetsOutput, bool) bool) error {
	return DescribeFleetsPagesWithContext(context.Background(), conn, input, fn)
}

func DescribeFleetsPagesWithContext(ctx context.Context, conn *appstream.AppStream, input *appstream.DescribeFleetsInput, fn func(*appstream.DescribeFleetsOutput, bool) bool) error {
	for {
		output, err := conn.DescribeFleetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}

func DescribeImageBuildersPages(conn *appstream.AppStream, input *appstream.DescribeImageBuildersInput, fn func(*appstream.DescribeImageBuildersOutput, bool) bool) error {
	return DescribeImageBuildersPagesWithContext(context.Background(), conn, input, fn)
}

func DescribeImageBuildersPagesWithContext(ctx context.Context, conn *appstream.AppStream, input *appstream.DescribeImageBuildersInput, fn func(*appstream.DescribeImageBuildersOutput, bool) bool) error {
	for {
		output, err := conn.DescribeImageBuildersWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}

func DescribeStacksPages(conn *appstream.AppStream, input *appstream.DescribeStacksInput, fn func(*appstream.DescribeStacksOutput, bool) bool) error {
	return DescribeStacksPagesWithContext(context.Background(), conn, input, fn)
}

func DescribeStacksPagesWithContext(ctx context.Context, conn *appstream.AppStream, input *appstream.DescribeStacksInput, fn func(*appstream.DescribeStacksOutput, bool) bool) error {
	for {
		output, err := conn.DescribeStacksWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
