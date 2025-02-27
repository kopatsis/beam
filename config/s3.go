package config

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadToS3(s3Client *s3.S3, filename string, file []byte) (string, error) {
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(REVIEW_BUCKET_NAME),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(file),
	})
	if err != nil {
		return "", err
	}

	fileURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", REVIEW_BUCKET_NAME, filename)
	return fileURL, nil
}

func DeleteFromS3(s3Client *s3.S3, fileURL string) error {
	if !strings.Contains(fileURL, fmt.Sprintf("https://%s.s3.amazonaws.com/", REVIEW_BUCKET_NAME)) {
		return fmt.Errorf("invalid URL format")
	}

	fileURL = strings.Replace(fileURL, fmt.Sprintf("https://%s.s3.amazonaws.com/", REVIEW_BUCKET_NAME), "", 1)

	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(REVIEW_BUCKET_NAME),
		Key:    aws.String(fileURL),
	})
	if err != nil {
		return err
	}
	return nil
}
