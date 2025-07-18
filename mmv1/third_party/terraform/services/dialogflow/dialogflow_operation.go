package dialogflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-provider-google/google/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google/google/transport"
)

type DialogflowOperationWaiter struct {
	Config    *transport_tpg.Config
	UserAgent string
	Project   string
	tpgresource.CommonOperationWaiter
}

func (w *DialogflowOperationWaiter) QueryOp() (interface{}, error) {
	if w == nil {
		return nil, fmt.Errorf("Cannot query operation, it's unset or nil.")
	}
	// Returns the proper get.
	location := ""
	if parts := regexp.MustCompile(`locations\/([^\/]*)\/`).FindStringSubmatch(w.CommonOperationWaiter.Op.Name); parts != nil {
		location = parts[1]
	} else {
		return nil, fmt.Errorf(
			"Saw %s when the op name is expected to contains location %s",
			w.CommonOperationWaiter.Op.Name,
			"projects/{{project}}/locations/{{location}}/...",
		)
	}

	url := fmt.Sprintf("https://%s-dialogflow.googleapis.com/v2/%s", location, w.CommonOperationWaiter.Op.Name)

	return transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
		Config:    w.Config,
		Method:    "GET",
		Project:   w.Project,
		RawURL:    url,
		UserAgent: w.UserAgent,
	})
}

func createDialogflowWaiter(config *transport_tpg.Config, op map[string]interface{}, project, activity, userAgent string) (*DialogflowOperationWaiter, error) {
	w := &DialogflowOperationWaiter{
		Config:    config,
		UserAgent: userAgent,
		Project:   project,
	}
	if err := w.CommonOperationWaiter.SetOp(op); err != nil {
		return nil, err
	}
	return w, nil
}

// nolint: deadcode,unused
func DialogflowOperationWaitTimeWithResponse(config *transport_tpg.Config, op map[string]interface{}, response *map[string]interface{}, project, activity, userAgent string, timeout time.Duration) error {
	w, err := createDialogflowWaiter(config, op, project, activity, userAgent)
	if err != nil {
		return err
	}
	if err := tpgresource.OperationWait(w, activity, timeout, config.PollInterval); err != nil {
		return err
	}
	rawResponse := []byte(w.CommonOperationWaiter.Op.Response)
	if len(rawResponse) == 0 {
		return errors.New("`resource` not set in operation response")
	}
	return json.Unmarshal(rawResponse, response)
}

func DialogflowOperationWaitTime(config *transport_tpg.Config, op map[string]interface{}, project, activity, userAgent string, timeout time.Duration) error {
	if val, ok := op["name"]; !ok || val == "" {
		// This was a synchronous call - there is no operation to wait for.
		return nil
	}
	w, err := createDialogflowWaiter(config, op, project, activity, userAgent)
	if err != nil {
		// If w is nil, the op was synchronous.
		return err
	}
	return tpgresource.OperationWait(w, activity, timeout, config.PollInterval)
}
