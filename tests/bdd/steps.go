package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	lambdaHandler "github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/aws/lambda"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
)

type bddWorld struct {
	payload  map[string]any
	resp     interface{}
	ctrlStub lambdaControllerStub
	require  *require.Assertions
}

type lambdaControllerStub struct {
	resp []byte
	err  error
}

func (s lambdaControllerStub) ProcessVideo(ctx context.Context, input dto.ProcessVideoInput) ([]byte, error) {
	if s.resp != nil || s.err != nil {
		return s.resp, s.err
	}
	// default body
	return []byte(`{"success":true,"message":"ok"}`), nil
}

func (w *bddWorld) iHaveLambdaEventWithVideoKey(key string) error {
	w.payload = map[string]any{"video_key": key}
	return nil
}

func (w *bddWorld) theControllerReturnsSuccess(frameCount int, outputKey string) error {
	body := fmt.Sprintf(`{"success":true,"message":"ok","frame_count":%d,"output_key":"%s"}`, frameCount, outputKey)
	w.ctrlStub.resp = []byte(body)
	w.ctrlStub.err = nil
	return nil
}

func (w *bddWorld) iInvokeTheLambdaHandler() error {
	h := lambdaHandler.NewHandler(w.ctrlStub, logger.NewSlogLogger())
	b, _ := json.Marshal(w.payload)
	resp, err := h.Handle(context.Background(), b)
	w.resp = resp
	return err
}

func (w *bddWorld) theResponseStatusCodeIs(code int) error {
	m := map[string]any{}
	b, _ := json.Marshal(w.resp)
	_ = json.Unmarshal(b, &m)
	if int(m["statusCode"].(float64)) != code {
		return fmt.Errorf("expected status %d, got %v", code, m["statusCode"])
	}
	return nil
}

func (w *bddWorld) theResponseJSONHasFieldEqualTo(field string, value string) error {
	m := map[string]any{}
	b, _ := json.Marshal(w.resp)
	_ = json.Unmarshal(b, &m)
	body := map[string]any{}
	_ = json.Unmarshal([]byte(m["body"].(string)), &body)

	// Trim surrounding quotes if present
	value = strings.Trim(value, `"'`)

	// Normalize expected value
	var expectedAny any
	switch strings.ToLower(value) {
	case "true":
		expectedAny = true
	case "false":
		expectedAny = false
	default:
		if iv, e := strconv.Atoi(value); e == nil {
			expectedAny = iv
		} else {
			expectedAny = value
		}
	}

	switch ev := expectedAny.(type) {
	case int:
		got := int(body[field].(float64))
		if got != ev {
			return fmt.Errorf("%s expected %d got %d", field, ev, got)
		}
	case string:
		got := body[field].(string)
		if got != ev {
			return fmt.Errorf("%s expected %s got %s", field, ev, got)
		}
	case bool:
		got := body[field].(bool)
		if got != ev {
			return fmt.Errorf("%s expected %v got %v", field, ev, got)
		}
	default:
		return fmt.Errorf("unsupported value type")
	}
	return nil
}

type testingT struct{}

func (t *testingT) Errorf(format string, args ...interface{}) {}
func (t *testingT) FailNow()                                  {}
