package grpc

type ShutdownService interface {
	Stop()
	GracefulStop()
}
