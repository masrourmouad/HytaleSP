msbuild Aurora\Aurora.slnx
go generate
go build -ldflags="-H windowsgui"