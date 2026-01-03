module github.com/moby/moby/client

go 1.24.0

require (
	github.com/containerd/errdefs v1.0.0
	github.com/containerd/errdefs/pkg v0.3.0
	github.com/distribution/reference v0.6.0
	github.com/docker/go-connections v0.6.0
	github.com/docker/go-units v0.5.0
	github.com/moby/moby/api v1.53.0-rc.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	gotest.tools/v3 v3.5.2
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/moby/moby/api => ../api
