package integrations

import "github.com/hatchet-dev/hatchet-workflows/pkg/workflows/types"

type Integration interface {
	GetId() string
	Actions() []string
	PerformAction(action types.Action, data map[string]interface{}) (map[string]interface{}, error)
}
