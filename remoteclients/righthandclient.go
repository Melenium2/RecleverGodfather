package remoteclients

import (
	"RecleverGodfather/grandlog"
	"github.com/Melenium2/RecleverRightHand/logic"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"time"
)

func NewRightHandClient(client consul.Client, logger grandlog.GrandLogger) http.Handler {
	var (
		tags        = []string{}
		passingOnly = true
		endpoints   = &logic.Endpoints{}
		instancer   = consul.NewInstancer(client, logger, "Right Hand", tags, passingOnly)
	)
	{
		factory := righthandFactory(logic.MakeStartEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.StartEndpoint = retry
	}
	{
		factory := righthandFactory(logic.MakeTerminateEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.TerminateEndpoint = retry
	}

	return logic.NewHttpTransport(endpoints, logger)
}

func righthandFactory(make func(s logic.Service) endpoint.Endpoint) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		client := logic.NewGrpcClient(conn)

		return make(client), conn, err
	}
}
