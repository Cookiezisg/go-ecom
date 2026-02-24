package file

import "github.com/zeromicro/go-zero/zrpc"

type config struct {
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
