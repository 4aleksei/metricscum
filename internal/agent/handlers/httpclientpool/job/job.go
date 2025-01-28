package job

import (
	"github.com/4aleksei/metricscum/internal/common/models"
)

type (
	JobID uint64

	Result struct {
		Result int
		Err    error
		ID     JobID
	}

	Job struct {
		ID    JobID
		Value []models.Metrics
	}
)
