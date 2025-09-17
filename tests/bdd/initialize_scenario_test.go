package tests

import (
	"context"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	w := &bddWorld{}
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		w.require = require.New(&testingT{})
		*w = bddWorld{}
		return c, nil
	})
	ctx.Step(`^a lambda event with video_key "([^"]*)"$`, w.iHaveLambdaEventWithVideoKey)
	ctx.Step(`^a lambda event with video_key "([^"]*)" and configuration frame_rate ([\d\.]+) and output_format "([^"]*)"$`, w.iHaveLambdaEventWithVideoKeyAndConfiguration)
	ctx.Step(`^the controller returns success with frame_count (\d+) and output_key "([^"]*)"$`, w.theControllerReturnsSuccess)
	ctx.Step(`^the controller returns success with frame_count (\d+) and output_key "([^"]*)" and hash "([^"]*)"$`, w.theControllerReturnsSuccessWithHash)
	ctx.Step(`^the controller returns success with frame_rate sanitized to ([\d\.]+) and format "([^"]*)" and frame_count (\d+)$`, w.theControllerReturnsSuccessWithSanitizedConfig)
	ctx.Step(`^I invoke the lambda handler$`, w.iInvokeTheLambdaHandler)
	ctx.Step(`^the response statusCode is (\d+)$`, w.theResponseStatusCodeIs)
	ctx.Step(`^the response JSON has field "([^"]*)" equal to (.*)$`, w.theResponseJSONHasFieldEqualTo)
	ctx.Step(`^the response JSON has field "([^"]*)" contains "([^"]*)"$`, w.theResponseJSONHasFieldContains)
	ctx.Step(`^the response JSON has field "([^"]*)" is not empty$`, w.theResponseJSONHasFieldIsNotEmpty)
}
