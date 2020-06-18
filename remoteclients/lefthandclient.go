package remoteclients

import (
	"RecleverGodfather/grandlog"
	murlog "github.com/Melenium2/Murlog"
	"github.com/Melenium2/RecleverLeftHand/logic"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"time"
)

func NewLeftHandClient(client consul.Client, logger grandlog.GrandLogger) http.Handler {
	var (
		tags        = []string{}
		passingOnly = true
		endpoints   = &logic.Endpoints{}
		instancer   = consul.NewInstancer(client, logger, "Left Hand", tags, passingOnly)
	)
	{
		factory := lefthandFactory(logic.MakeStartUpdatingEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.StartUpdatingEndpoint = retry
	}

	return logic.NewHttpTransport(endpoints, murlog.NewNopLogger())
}

func lefthandFactory(make func(s logic.Service) endpoint.Endpoint) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		client := logic.NewGrpcClient(conn)

		return make(client), conn, err
	}
}
