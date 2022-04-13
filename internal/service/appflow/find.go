package appflow

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindFlowByID(ctx context.Context, conn *appflow.Appflow, id string) (*appflow.Flow, error) {
	in := &appflow.GetFlowInput{
		Id: aws.String(id),
	}

	out, err := conn.GetFlowWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, appflow.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Flow == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Flow, nil
}
