package s3

import (
	"bytes"
	"context"
	"crypto/tls"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Storage interface {
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)
	ListObjects(ctx context.Context, prefix string) ([]*Object, error)

	GetObject(ctx context.Context, key string) ([]byte, error)
	PutObject(ctx context.Context, key string, object io.Reader, length int64, contentType string) (minio.UploadInfo, error)
	FPutObject(ctx context.Context, key string, path string, contentType string) (minio.UploadInfo, error)

	AddDirectory(ctx context.Context, path string) (minio.UploadInfo, error)

	GetLink(key string) string
	PresignedPutObject(ctx context.Context, key string, expires time.Duration) (*url.URL, error)

	RemoveObject(ctx context.Context, objectName string) error
}

type minioStorage struct {
	s3     *minio.Client
	bucket string
	url    *url.URL
}

type Object struct {
	Key         string
	Path        string
	Size        int64
	ContentType string
	UpdatedAt   time.Time
}

func newS3(logger *zap.Logger, config *Config, ctx context.Context) (Storage, error) {
	s := &minioStorage{}
	u, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}
	s.url = u
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = u.Scheme == "https"

	var transport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig,
	}
	minioClient, err := minio.New(u.Host, &minio.Options{
		Creds:        credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure:       u.Scheme == "https",
		Region:       config.Region,
		BucketLookup: minio.BucketLookupAuto,
		Transport:    transport,
	})

	if err != nil {
		return nil, err
	}
	s.s3 = minioClient
	s.bucket = config.Bucket

	return s, nil
}

func (s *minioStorage) GetClient() *minio.Client {
	return s.s3
}

func (s *minioStorage) AddDirectory(ctx context.Context, path string) (minio.UploadInfo, error) {
	return s.s3.PutObject(ctx, s.bucket, path+"/.keep", nil, 0, minio.PutObjectOptions{})
}

func (s *minioStorage) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	buckets, err := s.s3.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}
	return buckets, nil
}

func (s *minioStorage) GetLink(key string) string {
	if key == "" {
		return ""
	}
	sb := bytes.NewBufferString("")
	sb.WriteString(s.url.String())
	sb.WriteString("/")
	sb.WriteString(s.bucket)
	sb.WriteString("/")
	sb.WriteString(key)

	return sb.String()
}

func (s *minioStorage) PresignedPutObject(ctx context.Context, key string, expires time.Duration) (*url.URL, error) {
	return s.s3.PresignedPutObject(ctx, s.bucket, key, expires)
}

func (s *minioStorage) PutObject(ctx context.Context, key string, object io.Reader, length int64, contentType string) (minio.UploadInfo, error) {
	return s.s3.PutObject(ctx, s.bucket, key, object, length, minio.PutObjectOptions{ContentType: contentType})
}

func (s *minioStorage) FPutObject(ctx context.Context, key string, path string, contentType string) (minio.UploadInfo, error) {
	return s.s3.FPutObject(ctx, s.bucket, key, path, minio.PutObjectOptions{ContentType: contentType})
}

func (s *minioStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	result, err := s.s3.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(result)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *minioStorage) ListObjects(ctx context.Context, prefix string) ([]*Object, error) {
	opts := minio.ListObjectsOptions{
		Recursive: false,
		Prefix:    prefix,
	}
	var objs []*Object
	for object := range s.s3.ListObjects(ctx, s.bucket, opts) {
		if object.Err != nil {
			return nil, object.Err
		}
		if strings.HasPrefix(strings.Replace(object.Key, prefix, "", -1), ".") {
			continue
		}
		objs = append(objs, &Object{
			Key:         strings.Replace(object.Key, prefix, "", -1),
			Path:        object.Key,
			Size:        object.Size,
			ContentType: object.ContentType,
			UpdatedAt:   object.LastModified,
		})
	}
	return objs, nil
}

func (s *minioStorage) RemoveObject(ctx context.Context, objectName string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	return s.s3.RemoveObject(ctx, s.bucket, objectName, opts)
}
