package server

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"
	types "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// networkNotReadyReason is the reason reported when network is not ready.
const networkNotReadyReason = "NetworkPluginNotReady"

// Status returns the status of the runtime
func (s *Server) Status(ctx context.Context, req *types.StatusRequest) (*types.StatusResponse, error) {
	runtimeCondition := &types.RuntimeCondition{
		Type:   types.RuntimeReady,
		Status: true,
	}
	networkCondition := &types.RuntimeCondition{
		Type:   types.NetworkReady,
		Status: true,
	}

	if err := s.config.CNIPluginReadyOrError(); err != nil {
		networkCondition.Status = false
		networkCondition.Reason = networkNotReadyReason
		networkCondition.Message = fmt.Sprintf("Network plugin returns error: %v", err)
	}

	resp := &types.StatusResponse{
		Status: &types.RuntimeStatus{
			Conditions: []*types.RuntimeCondition{
				runtimeCondition,
				networkCondition,
			},
		},
	}
	if req.Verbose {
		info, err := s.createRuntimeInfo()
		if err != nil {
			return nil, fmt.Errorf("creating runtime info: %w", err)
		}
		resp.Info = info
	}
	return resp, nil
}

func (s *Server) createRuntimeInfo() (map[string]string, error) {
	config := map[string]interface{}{
		"sandboxImage": s.config.ImageConfig.PauseImage,
	}
	bytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}
	return map[string]string{"config": string(bytes)}, nil
}
