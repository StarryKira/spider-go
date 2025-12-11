package evaluation

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetEvaluationInfo(ctx context.Context, uid int)
}

func (h *Handler) GetEvaluationInfo(ctx context.Context, uid int) {

}
