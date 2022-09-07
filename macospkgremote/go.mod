module github.com/gohugoio/hugoreleaser-archive-plugins/macospkgremote

go 1.19

require (
	github.com/bep/helpers v0.4.0
	github.com/bep/s3rpc v0.2.0
	github.com/gohugoio/hugoreleaser-archive-plugins/macospkg v0.0.0-00010101000000-000000000000
	github.com/gohugoio/hugoreleaser-plugins-api v0.7.0
	golang.org/x/sync v0.0.0-20220907140024-f12130a52804
)

require (
	github.com/aws/aws-sdk-go v1.44.94 // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.14 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.18 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.31 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.18.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.27.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.19.8 // indirect
	github.com/aws/smithy-go v1.13.2 // indirect
	github.com/bep/awscreate v0.1.0 // indirect
	github.com/bep/awscreate/s3rpccreate v0.2.0 // indirect
	github.com/bep/buildpkg v0.1.0 // indirect
	github.com/bep/execrpc v0.7.1 // indirect
	github.com/bep/macosnotarylib v0.1.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.3-0.20220820150458-bfea432b1a9d // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
)

replace github.com/gohugoio/hugoreleaser-archive-plugins/macospkg => ../macospkg
