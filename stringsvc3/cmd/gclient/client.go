package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Client using go-kit http transport with a fallback in case of error for uppercase
func main() {
	fUpperCase := flag.Bool("uppercase", false, "Call uppercase operation")
	fCount := flag.Bool("count", false, "Call count operation")
	flag.Parse()
	if flag.NFlag() != 1 {
		flag.Usage()
		return
	}
	reqText := strings.Join(flag.Args(), " ")
	if reqText == "" {
		reqText = "Sample Text"
	}

	var err error
	if *fUpperCase {
		var respText string
		respText, err = callUpperCase(reqText)
		if err == nil {
			fmt.Printf("Succcess: uppercase result %v\n", respText)
		}

	} else if *fCount {
		var respCount int
		respCount, err = callCount(reqText)
		if err == nil {
			fmt.Printf("Succcess: count result %v\n", respCount)
		}
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func callCount(reqText string) (int, error) {
	tHost := "http://localhost:8080/"
	tResource, err := url.Parse(tHost + "count")
	if err != nil {
		panic(err)
	}
	client := httptransport.NewClient(
		http.MethodGet,
		tResource,
		httptransport.EncodeJSONRequest,
		decodeCountResponse,
	).Endpoint()

	rRaw, err := client(context.Background(), countRequest{S: reqText})
	if err != nil {
		return 0, fmt.Errorf("client failed with error %w", err)
	}

	return rRaw.(countResponse).V, nil
}

func decodeCountResponse(_ context.Context, r *http.Response) (interface{}, error) {
	var response countResponse
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func callUpperCase(reqText string) (string, error) {
	getEndpoint := func(port string) endpoint.Endpoint {
		tResource, err := url.Parse("http://localhost:" + port + "/uppercase")
		if err != nil {
			panic(err)
		}

		client := httptransport.NewClient(
			http.MethodGet,
			tResource,
			httptransport.EncodeJSONRequest,
			decodeUppercaseResponse,
		).Endpoint()
		return client
	}

	endpoint := fallbackRetry{
		main:      getEndpoint("8080"),
		fallbacks: []endpoint.Endpoint{getEndpoint("8081"), getEndpoint("8082")},
	}.Endpoint()

	rRaw, err := endpoint(context.Background(), uppercaseRequest{S: reqText})
	if err != nil {
		return "", fmt.Errorf("client failed with error %w", err)
	}

	resp := rRaw.(uppercaseResponse)
	if resp.Err != "" {
		return "", fmt.Errorf("received error from server %v", resp.Err)
	}

	return resp.V, nil
}

func decodeUppercaseResponse(_ context.Context, r *http.Response) (interface{}, error) {
	var response uppercaseResponse
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

// fallbackRetry provides an Endpoint with a retry behaviour after the main endpoint fails.
// The Endpoint() returned will always try the main endpoint first and then retry on fallback endpoints
// until a success response is received or all fallback endpoints are exhausted.
type fallbackRetry struct {
	main      endpoint.Endpoint
	fallbacks []endpoint.Endpoint
}

func (fr fallbackRetry) Endpoint() endpoint.Endpoint {
	if len(fr.fallbacks) == 0 {
		panic("fallbacks not filled")
	}

	balancer := lb.NewRoundRobin(sd.FixedEndpointer(fr.fallbacks))
	retry := lb.Retry(len(fr.fallbacks), time.Duration(len(fr.fallbacks))*time.Second, balancer)

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		resp, err := fr.main(ctx, request)
		if err != nil {
			log.Printf("Retry to fallbacks after main endpoint failed with: %s", err)
			return retry(ctx, request)
		} else {
			return resp, err
		}
	}
}

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"`
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}
