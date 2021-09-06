package sources

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jszwec/s3fs"
	"github.com/linefusion/pages/pkg/fsh"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type S3Source struct {
	BaseSource
	Root            hcl.Expression `hcl:"root,optional"`
	Bucket          string         `hcl:"bucket,attr"`
	AccessKeyID     string         `hcl:"access_key_id,attr"`
	AccessSecretKey string         `hcl:"access_secret_key,attr"`
	Endpoint        string         `hcl:"endpoint,attr"`
	Region          string         `hcl:"region,optional"`
}

func (source *S3Source) CreateFs(context hcl.EvalContext, request *http.Request) (fs.FS, error) {
	rootDir, err := os.Getwd()
	if err != nil {
		rootDir = "/"
	}

	root, err := evaluate(context, source.Root, cty.StringVal(rootDir))
	if err != nil {
		return nil, err
	}

	if !root.IsNull() {
		rootDir = root.AsString()
	}

	session, _ := session.NewSession(&aws.Config{
		Region:           &source.Region,
		Endpoint:         &source.Endpoint,
		Credentials:      credentials.NewStaticCredentials(source.AccessKeyID, source.AccessSecretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})

	return fsh.Normalize(s3fs.New(s3.New(session), source.Bucket), fsh.NormalizeOptions{
		TrimLeadingSlash: true,
		Separator:        fsh.ForwardSeparator,
		Prefix:           rootDir,
	}), nil
}

func (source *S3Source) Configure() {
	source.CacheKeys().UseExpression(source.Root)
	if source.Region == "" {
		source.Region = "us-east-1"
	}
}

func init() {
	Register("s3", &S3Source{})
}
