package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	deliveryV1 "github.com/Venamuss/webhook-dispatcher/api/proto/delivery/v1"
	"github.com/Venamuss/webhook-dispatcher/internal/subscription"
)

type ReporterServer struct {
	service *subscription.Service
	deliveryV1.UnimplementedDeliveryReporterServiceServer
}

func NewReporterServer(service *subscription.Service) *ReporterServer {
	return &ReporterServer{
		service: service,
	}
}

func (s *ReporterServer) ReportAttempt(ctx context.Context, req *deliveryV1.ReportAttemptRequest) (*deliveryV1.ReportAttemptResponse, error) {
	if req.GetEventId() == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	if _, err := uuid.Parse(req.GetEndpointId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid endpoint_id")
	}

	domainStatus := mapProtoStatus(req.GetStatus())
	if domainStatus == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	attempt := &subscription.DeliveryAttempt{
		EventID:         req.GetEventId(),
		EndpointID:      req.GetEndpointId(),
		AttemptNumber:   int(req.GetAttemptNumber()),
		Status:          domainStatus,
		HTTPStatusCode:  int(req.GetHttpStatusCode()),
		ExecutionTimeMS: int(req.GetExecutionTimeMs()),
		ResponseBody:    req.GetResponseBody(),
		ErrorMessage:    req.GetErrorMessage(),
		TraceID:         req.GetTraceId(),
	}

	if err := s.service.RecordDeliveryAttempt(ctx, attempt); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &deliveryV1.ReportAttemptResponse{Acknowledged: true}, nil
}

func mapProtoStatus(s deliveryV1.DeliveryStatus) subscription.Status {
	switch s {
	case deliveryV1.DeliveryStatus_DELIVERY_STATUS_SUCCESS:
		return subscription.StatusSuccess
	case deliveryV1.DeliveryStatus_DELIVERY_STATUS_FAILED:
		return subscription.StatusFailed
	case deliveryV1.DeliveryStatus_DELIVERY_STATUS_RETRYING:
		return subscription.StatusRetrying
	case deliveryV1.DeliveryStatus_DELIVERY_STATUS_DEAD_LETTER:
		return subscription.StatusDeadLetter
	default:
		return ""
	}
}
