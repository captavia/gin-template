package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/samber/do/v2"

	"template/config"
)

// 领域错误定义
var (
	ErrObjectNotFound      = errors.New("s3: object not found")
	ErrBucketNotConfigured = errors.New("s3: bucket not configured")
	ErrInvalidPath         = errors.New("s3: invalid object path")
	ErrInvalidExpiration   = errors.New("s3: expiration must be positive and not exceed 7 days")
)

// ObjectInfo 对象信息结构体
type ObjectInfo struct {
	Bucket       string            `json:"bucket"`
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ETag         string            `json:"etag"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	UserMetadata map[string]string `json:"user_metadata,omitempty"`
}

// StorageOptions 操作配置选项
type StorageOptions struct {
	Bucket       string
	ContentType  string
	UserMetadata map[string]string
	QueryParams  url.Values
}

type StorageOption func(*StorageOptions)

// WithBucket 指定自定义 Bucket 覆盖默认配置
func WithBucket(bucket string) StorageOption {
	return func(o *StorageOptions) {
		o.Bucket = strings.TrimSpace(bucket)
	}
}

// WithContentType 指定 Content-Type
func WithContentType(contentType string) StorageOption {
	return func(o *StorageOptions) {
		o.ContentType = strings.TrimSpace(contentType)
	}
}

// WithUserMetadata 指定自定义用户元数据
func WithUserMetadata(metadata map[string]string) StorageOption {
	return func(o *StorageOptions) {
		o.UserMetadata = metadata
	}
}

// WithQueryParams 为预签名链接指定附加查询参数
func WithQueryParams(params url.Values) StorageOption {
	return func(o *StorageOptions) {
		o.QueryParams = params
	}
}

// S3Service 提供 S3 对象存储业务逻辑与能力
type S3Service struct {
	client *minio.Client
	cfg    *config.Config
}

// NewS3Service DI 依赖注入构造函数
func NewS3Service(i do.Injector) (*S3Service, error) {
	return &S3Service{
		client: do.MustInvoke[*minio.Client](i),
		cfg:    do.MustInvoke[*config.Config](i),
	}, nil
}

// Client 返回底层 MinIO Client 实例
func (s *S3Service) Client() *minio.Client {
	return s.client
}

// Put 写入数据流到 S3
func (s *S3Service) Put(ctx context.Context, path string, reader io.Reader, size int64, opts ...StorageOption) (*ObjectInfo, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return nil, err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return nil, err
	}

	contentType := options.ContentType
	if contentType == "" {
		contentType = detectContentType(cleanPath, nil)
	}

	putOpts := minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: options.UserMetadata,
	}

	info, err := s.client.PutObject(ctx, bucket, cleanPath, reader, size, putOpts)
	if err != nil {
		return nil, err
	}

	return &ObjectInfo{
		Bucket:       bucket,
		Key:          cleanPath,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  contentType,
		LastModified: time.Now(),
		UserMetadata: options.UserMetadata,
	}, nil
}

// PutBytes 写入字节切片到 S3
func (s *S3Service) PutBytes(ctx context.Context, path string, data []byte, opts ...StorageOption) (*ObjectInfo, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return nil, err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return nil, err
	}

	contentType := options.ContentType
	if contentType == "" {
		header := data
		if len(header) > 512 {
			header = header[:512]
		}
		contentType = detectContentType(cleanPath, header)
	}

	putOpts := minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: options.UserMetadata,
	}

	info, err := s.client.PutObject(ctx, bucket, cleanPath, bytes.NewReader(data), int64(len(data)), putOpts)
	if err != nil {
		return nil, err
	}

	return &ObjectInfo{
		Bucket:       bucket,
		Key:          cleanPath,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  contentType,
		LastModified: time.Now(),
		UserMetadata: options.UserMetadata,
	}, nil
}

// Stat 根据路径查询对象元信息
func (s *S3Service) Stat(ctx context.Context, path string, opts ...StorageOption) (*ObjectInfo, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return nil, err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return nil, err
	}

	info, err := s.client.StatObject(ctx, bucket, cleanPath, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.StatusCode == http.StatusNotFound {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	return &ObjectInfo{
		Bucket:       bucket,
		Key:          cleanPath,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
		UserMetadata: info.UserMetadata,
	}, nil
}

// Exists 根据路径判断对象是否存在
func (s *S3Service) Exists(ctx context.Context, path string, opts ...StorageOption) (bool, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return false, err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return false, err
	}

	_, err = s.client.StatObject(ctx, bucket, cleanPath, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}

	errResp := minio.ToErrorResponse(err)
	if errResp.Code == "NoSuchKey" || errResp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, err
}

// GetPresignedURL 获取对象的临时预签名访问链接
func (s *S3Service) GetPresignedURL(ctx context.Context, path string, expires time.Duration, opts ...StorageOption) (string, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return "", err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return "", err
	}

	if expires <= 0 {
		expires = 1 * time.Hour
	}
	if expires > 7*24*time.Hour {
		return "", ErrInvalidExpiration
	}

	reqParams := make(url.Values)
	if options.QueryParams != nil {
		for k, v := range options.QueryParams {
			reqParams[k] = v
		}
	}

	u, err := s.client.PresignedGetObject(ctx, bucket, cleanPath, expires, reqParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// GetPreviewURL 获取文件在线预览链接
// 通过显式设置 response-content-disposition=inline 以及正确的 response-content-type，使浏览器可直接展示图片、PDF、音视频等
func (s *S3Service) GetPreviewURL(ctx context.Context, path string, expires time.Duration, opts ...StorageOption) (string, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return "", err
	}

	// 合并用户传入的查询参数
	options := s.applyOptions(opts...)
	reqParams := make(url.Values)
	if options.QueryParams != nil {
		for k, v := range options.QueryParams {
			reqParams[k] = v
		}
	}

	// 强制以 inline 形式打开
	reqParams.Set("response-content-disposition", "inline")
	if options.ContentType != "" {
		reqParams.Set("response-content-type", options.ContentType)
	} else {
		if guessed := detectContentType(cleanPath, nil); guessed != "" {
			reqParams.Set("response-content-type", guessed)
		}
	}

	// 重新构建选项：保留 bucket/contentType/metadata，使用新的 queryParams
	finalOpts := []StorageOption{WithQueryParams(reqParams)}
	if options.Bucket != "" && options.Bucket != s.cfg.S3.Bucket {
		finalOpts = append(finalOpts, WithBucket(options.Bucket))
	}
	if options.UserMetadata != nil {
		finalOpts = append(finalOpts, WithUserMetadata(options.UserMetadata))
	}

	return s.GetPresignedURL(ctx, cleanPath, expires, finalOpts...)
}

// GetDownloadURL 获取文件下载链接
// 通过显式设置 response-content-disposition=attachment; filename="..." 强制触发浏览器下载
func (s *S3Service) GetDownloadURL(ctx context.Context, path string, filename string, expires time.Duration, opts ...StorageOption) (string, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return "", err
	}

	if filename == "" {
		filename = filepath.Base(cleanPath)
	}

	// 合并用户传入的查询参数
	options := s.applyOptions(opts...)
	reqParams := make(url.Values)
	if options.QueryParams != nil {
		for k, v := range options.QueryParams {
			reqParams[k] = v
		}
	}

	disposition := fmt.Sprintf("attachment; filename=%q", filename)
	reqParams.Set("response-content-disposition", disposition)

	// 重新构建选项：保留 bucket/contentType/metadata，使用新的 queryParams
	finalOpts := []StorageOption{WithQueryParams(reqParams)}
	if options.Bucket != "" && options.Bucket != s.cfg.S3.Bucket {
		finalOpts = append(finalOpts, WithBucket(options.Bucket))
	}
	if options.ContentType != "" {
		finalOpts = append(finalOpts, WithContentType(options.ContentType))
	}
	if options.UserMetadata != nil {
		finalOpts = append(finalOpts, WithUserMetadata(options.UserMetadata))
	}

	return s.GetPresignedURL(ctx, cleanPath, expires, finalOpts...)
}

// GetPublicURL 获取对外公开直链（适用于公共 Bucket 或使用 CDN/Nginx 反代的场景）
func (s *S3Service) GetPublicURL(path string, opts ...StorageOption) (string, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return "", err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(s.cfg.S3.Endpoint, "/")
	if endpoint == "" {
		return "", errors.New("s3: endpoint not configured")
	}

	scheme := "http"
	if s.cfg.S3.SSL {
		scheme = "https"
	}

	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return fmt.Sprintf("%s/%s/%s", endpoint, bucket, cleanPath), nil
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint, bucket, cleanPath), nil
}

// Get 获取文件流，调用方负责关闭返回的 *minio.Object
func (s *S3Service) Get(ctx context.Context, path string, opts ...StorageOption) (*minio.Object, error) {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return nil, err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return nil, err
	}

	return s.client.GetObject(ctx, bucket, cleanPath, minio.GetObjectOptions{})
}

// GetBytes 获取文件内容为字节切片
func (s *S3Service) GetBytes(ctx context.Context, path string, opts ...StorageOption) ([]byte, error) {
	obj, err := s.Get(ctx, path, opts...)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.StatusCode == http.StatusNotFound {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	return data, nil
}

// Delete 删除指定路径的文件
func (s *S3Service) Delete(ctx context.Context, path string, opts ...StorageOption) error {
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return err
	}
	options := s.applyOptions(opts...)
	bucket, err := s.resolveBucket(options.Bucket)
	if err != nil {
		return err
	}

	return s.client.RemoveObject(ctx, bucket, cleanPath, minio.RemoveObjectOptions{})
}

// 辅助方法

func (s *S3Service) applyOptions(opts ...StorageOption) StorageOptions {
	options := StorageOptions{
		Bucket: s.cfg.S3.Bucket,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func (s *S3Service) cleanPath(rawPath string) (string, error) {
	clean := strings.TrimPrefix(strings.TrimSpace(rawPath), "/")
	if clean == "" {
		return "", ErrInvalidPath
	}
	return clean, nil
}

func (s *S3Service) resolveBucket(bucket string) (string, error) {
	b := strings.TrimSpace(bucket)
	if b == "" {
		return "", ErrBucketNotConfigured
	}
	return b, nil
}

func detectContentType(path string, header []byte) string {
	ext := filepath.Ext(path)
	if ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}
	if len(header) > 0 {
		return http.DetectContentType(header)
	}
	return "application/octet-stream"
}
