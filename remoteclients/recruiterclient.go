package remoteclients

import (
	"RecleverGodfather/grandlog"
	"github.com/Melenium2/RecleverRecruiter/logic"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"time"
)

func NewRecruiterClient(client consul.Client, logger grandlog.GrandLogger) http.Handler {
	var (
		tags        = []string{}
		passingOnly = true
		endpoints   = &logic.Endpoints{}
		instancer   = consul.NewInstancer(client, logger, "recruiter", tags, passingOnly)
	)
	{
		factory := recruiterFactory(logic.MakeGetAccountEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.GetAccountEndpoint = retry
	}
	{
		factory := recruiterFactory(logic.MakeGetAccountNotEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.GetAccountNotEndpoint = retry
	}
	{
		factory := recruiterFactory(logic.MakeInsertNewEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.InsertNewEndpoint = retry
	}
	{
		factory := recruiterFactory(logic.MakeUpdateEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.UpdateEndpoint = retry
	}

	logger.Log("type", "[Info]", "action", "connect to recruiter")
	return logic.NewHttpTransport(endpoints, logger)
}

func recruiterFactory(make func(logic.Service) endpoint.Endpoint) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		client := logic.NewGrpcClient(conn)

		return make(client), conn, err
	}
}
