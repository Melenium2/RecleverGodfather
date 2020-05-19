package remoteclients

import (
	"RecleverGodfather/grandlog"
	guard "RecleverGodfather/proto"
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"time"
)

type endpoints struct {
	charts endpoint.Endpoint
}

func NewGuardClient(client consul.Client, logger grandlog.GrandLogger) http.Handler {
	var (
		tags        = []string{}
		passingOnly = true
		endpoints   = &endpoints{}
		instancer   = consul.NewInstancer(client, logger, "guard", tags, passingOnly)
	)
	{
		factory := guardFactory(makeTopChartEndpoint, logger)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(3, time.Second*180, balancer)
		endpoints.charts = retry
	}

	return createHttpTransport(endpoints)
}

func createHttpTransport(ep *endpoints) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeHttpError),
	}

	charts := kithttp.NewServer(
		ep.charts,
		httpDecodeRequest,
		httpEncodeResponse,
		opts...
	)

	r := mux.NewRouter()
	r.Handle("/apps", charts).Methods("GET")

	return accessControl(r)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func httpDecodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	log.Print("New request")
	return nil, nil
}

func httpEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return encodeHTTPResponse(ctx, http.StatusOK, w, response)
}

func makeTopChartEndpoint(client guard.GuardClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		cred := &guard.ChartsRequest{
			Cat:    "apps_topselling_paid_GAME",
			SubCat: "",
			Account: &guard.Account{
				Login:    "dd923633@gmail.com",
				Password: "s0EJ3xuPq",
				GsfId:    3650847969207086211,
				Token:    "wwfEo6-0oPWKjb3PWnHrvN06_5jc3XBwMvLwrQB3XL7m9bSxJZ_82AtsPCrInUIV5qlpiA.",
				Locale:   "en_US",
				Proxy: &guard.Proxy{
					Http:  "http://LQjdXZ:KJneZB@181.177.85.211:9900",
					Https: "https://LQjdXZ:KJneZB@181.177.85.211:9900",
					No:    "192.168.99.100",
				},
				Device: "whyred",
			},
		}
		resp, err := client.TopCharts(ctx, cred)
		if err != nil {
			return nil, err
		}

		return resp, err
	}
}

func guardFactory(makeEndpoint func(client guard.GuardClient) endpoint.Endpoint, logger grandlog.GrandLogger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		logger.Log("[Info]", "Instance trying to connect ", instance)
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		client := guard.NewGuardClient(conn)
		logger.Log("[Info]", "connection establish")
		return makeEndpoint(client), conn, err
	}
}
