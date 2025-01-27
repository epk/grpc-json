package main

import (
	"io"
	"net/http"

	"google.golang.org/grpc/codes"
	healthcheck "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// Provides JSON REST compatibility for health check service
type JSONHandler struct {
	client healthcheck.HealthClient
}

func NewJSONHandler(c healthcheck.HealthClient) http.Handler {
	handler := &JSONHandler{
		client: c,
	}

	mux := http.NewServeMux()
	// Add routes for health check service
	mux.HandleFunc(healthcheck.Health_Check_FullMethodName, handler.Check)
	mux.HandleFunc(healthcheck.Health_Watch_FullMethodName, handler.Watch)

	return mux
}

func (h *JSONHandler) Check(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Decode the JSON request
	var req healthcheck.HealthCheckRequest
	if err := protojson.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Call the gRPC client's Check method
	resp, err := h.client.Check(r.Context(), &req)
	if err != nil {
		httpError(w, err)
		return
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")

	respBytes, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Write(respBytes)
}

func (h *JSONHandler) Watch(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Decode the JSON request
	var req healthcheck.HealthCheckRequest
	if err := protojson.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Call the gRPC client's Watch method and stream responses
	stream, err := h.client.Watch(r.Context(), &req)
	if err != nil {
		httpError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json+stream")
	w.Header().Set("Transfer-Encoding", "chunked")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			httpError(w, err)
			return
		}

		respBytes, err := protojson.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}
		respBytes = append(respBytes, '\n')

		if _, err := w.Write(respBytes); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		flusher.Flush()
	}
}

func httpError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, "Unknown error", http.StatusInternalServerError)
		return
	}

	http.Error(w, st.Message(), HTTPStatusFromCode(st.Code()))
}

// Taken from https://github.com/grpc-ecosystem/grpc-gateway/blob/d65d53c586c2327c08990c9775e584544d3693a1/runtime/errors.go#L34-L77
func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
