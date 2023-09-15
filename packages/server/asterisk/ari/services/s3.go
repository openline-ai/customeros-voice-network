package services

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	awsSes "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/ini.v1"
	"io"
	"time"
)

type S3Client interface {
	GetFile(id string) (io.ReadCloser, error)
	GetUrl(id string) (string, error)
}

type S3ClientImpl struct {
	region string
	bucket string
}

func newS3Client(cfg *ini.File) S3Client {
	return &S3ClientImpl{
		region: cfg.Section("aws").Key("region").String(),
		bucket: cfg.Section("aws").Key("bucket").String(),
	}
}

func (ssi *S3ClientImpl) GetFile(id string) (io.ReadCloser, error) {
	session, err := awsSes.NewSession(&aws.Config{Region: aws.String(ssi.region)})
	if err != nil {
		return nil, fmt.Errorf("GetFile: Error creating session: %v", err)
	}
	svc := s3.New(session)

	// Get the object metadata to determine the file size and ETag
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(ssi.bucket),
		Key:    aws.String(id),
	})
	if err != nil {
		// Handle error
		return nil, fmt.Errorf("GetFile: Error getting object: %v", err)
	}
	return resp.Body, nil
}

func (ssi *S3ClientImpl) GetUrl(id string) (string, error) {
	session, err := awsSes.NewSession(&aws.Config{Region: aws.String(ssi.region)})
	if err != nil {
		return "", fmt.Errorf("GetFile: Error creating session: %v", err)
	}
	svc := s3.New(session)
	// Get the object metadata to determine the file size and ETag
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(ssi.bucket),
		Key:    aws.String(id),
	})
	urlStr, err := req.Presign(15 * time.Minute)

	if err != nil {
		return "", fmt.Errorf("GetUrl: Error signing the object: %v", err)
	}
	return urlStr, nil
}
