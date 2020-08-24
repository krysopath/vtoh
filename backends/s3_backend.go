//// +build backends/s3

package backends

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/yaml.v2"
)

type S3Backend struct {
	Bucket string
	Path   string
	Region string
}

func NewS3Client() (*s3.Client, context.Context) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err)
	}

	svc := s3.New(cfg)
	ctx := context.Background()
	return svc, ctx
}

func (backend S3Backend) Load() ([]byte, error) {
	svc, ctx := NewS3Client()
	req := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(backend.Bucket),
		Key:    aws.String(strings.Trim(backend.Path, "/")),
	})

	resp, err := req.Send(ctx)
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return content, nil
}

func (backend S3Backend) Save(data interface{}) (bool, error) {
	dataBytes, yamlErr := yaml.Marshal(data)
	if yamlErr != nil {
		panic(yamlErr)
	}
	reader := bytes.NewReader(dataBytes)
	svc, ctx := NewS3Client()
	req := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(backend.Bucket),
		Key:    aws.String(strings.Trim(backend.Path, "/")),
		Body:   reader,
	})
	_, reqErr := req.Send(ctx)
	if reqErr != nil {
		panic(reqErr)
	}
	return true, nil
}
