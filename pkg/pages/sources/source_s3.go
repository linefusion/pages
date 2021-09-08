package sources

import (
	"io/fs"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/linefusion/pages/pkg/iofs/s3"
	"github.com/valyala/fasthttp"

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

func (source *S3Source) CreateFs(request *fasthttp.Request, context hcl.EvalContext) (fs.FS, error) {
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

	return s3.New(awss3.New(session), source.Bucket, rootDir), nil
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
