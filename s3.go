package s3

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
)

type S3Config struct {
	Key, Secret, Region, Host, Bucket string
}

var S3 = &S3Config{}

func (a *S3Config) Init(key, secret, region, host, bucket string) {
	a.Key, a.Secret, a.Region, a.Host, a.Bucket = key, secret, region, host, bucket
}

func (a *S3Config) UploadFile(bucket, destination, imgType string, buffer []byte) (string, error) {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(a.Region),
		Credentials: credentials.NewStaticCredentials(a.Key, a.Secret, ""),
	})
	if err != nil {
		return "", err
	}
	if bucket == "" {
		bucket = a.Bucket
	}

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(destination),
		ACL:         aws.String("public-read"),
		Body:        bytes.NewReader(buffer),
		ContentType: aws.String(imgType),
	})
	return "https://" + bucket + ".s3." + a.Region + "." + a.Host + destination, err
}

func (a *S3Config) UploadJpeg(m image.Image, destination string) (string, error) {
	m = resize.Thumbnail(1028, 1028, m, resize.Lanczos3)
	buf, cnt := new(bytes.Buffer), 0
	for {
		if err := jpeg.Encode(buf, m, &jpeg.Options{Quality: 75}); err != nil {
			return "", err
		}
		if buf.Len() < 250000 || cnt > 10 {
			break
		}
		cnt++
	}

	return a.UploadFile("", destination, "jpeg", buf.Bytes())
}

func (a *S3Config) UploadBase64(base64Image, destination string) (string, error) {
	tmp := strings.Split(base64Image, ";base64,")
	if len(tmp) != 2 {
		return "", errors.New("cannot encode image")
	}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(tmp[1]))
	m, _, err := image.Decode(reader)
	if err != nil {
		return "", err
	}
	return a.UploadJpeg(m, destination)
}

func (a *S3Config) UploadPng(m image.Image, destination string) (string, error) {
	m = resize.Thumbnail(1028, 1028, m, resize.Lanczos3)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, m); err != nil {
		return "", err
	}
	return a.UploadFile("", destination, "png", buf.Bytes())
}
