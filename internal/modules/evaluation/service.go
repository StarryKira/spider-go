package evaluation

import "context"

type Service interface {
	GetEvaluationInfo(ctx context.Context, uid int)
}
