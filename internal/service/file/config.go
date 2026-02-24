package file

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcClientConf
	Storage StorageConfig
}

type StorageConfig struct {
	Type      string
	LocalPath string
	OSS       OSSConfig
}

type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}
