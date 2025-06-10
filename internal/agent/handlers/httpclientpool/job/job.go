package job

import (
	"github.com/4aleksei/metricscum/internal/common/models"
)

type (
	JobID uint64

	Result struct {
		Err    error
		Result int
		ID     JobID
	}

	Job struct {
		Value []models.Metrics
		ID    JobID
	}

	JobDone struct{}
)
