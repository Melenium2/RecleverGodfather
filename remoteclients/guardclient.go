package remoteclients

import (
	"RecleverGrandfather/grandlog"
	guard "RecleverGrandfather/proto"
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"io"
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
		retry := lb.Retry(3, time.Second * 60, balancer)
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
	return nil, nil
}

func httpEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	//resp := response.(*CreateResponse)
	//if resp.Err == "" {
	return encodeHTTPResponse(ctx, http.StatusOK, w, response)
	//}
	//encodeHttpError(ctx, getHTTPError(resp.Err), w)
	//return nil
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
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		c, closer, err := newGuardClient(conn)
		return makeEndpoint(c), closer, err
	}
}

func newGuardClient(c *grpc.ClientConn) (guard.GuardClient, *grpc.ClientConn, error) {
	return guard.NewGuardClient(c), c, nil
}

func encodeHttpError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//switch err {
	//case errBadRequest:
	//	w.WriteHeader(http.StatusBadRequest)
	//case errBadRoute:
	//	w.WriteHeader(http.StatusNotFound)
	//default:
	w.WriteHeader(http.StatusBadRequest)
	//}
	e := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func encodeHTTPResponse(_ context.Context, code int, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(code)
	if response == nil {
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
