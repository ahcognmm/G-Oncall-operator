package oncall_schema

import (
	oncallv1 "github.com/ahcogn/grafana-oncall-operator/api/v1"
)

type Integration struct {
	oncallv1.IntegrationSpec
	Link string `json:"link"`
}
