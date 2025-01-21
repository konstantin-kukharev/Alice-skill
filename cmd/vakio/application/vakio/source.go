package vakio

import (
	"context"
	"time"

	"github.com/konstantin-kukharev/Alice-skill/internal/logger"
)

type api interface {
	GetToken() string
}

type Source struct {
	l  *logger.Logger
	a  api
	pi time.Duration
}

func NewSourceApp(l *logger.Logger, a api, poolInterval time.Duration) *Source {
	return &Source{
		l:  l,
		a:  a,
		pi: poolInterval,
	}
}

func (s *Source) Run(ctx context.Context) error {
	return nil
}
