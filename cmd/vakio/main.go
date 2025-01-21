package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/konstantin-kukharev/Alice-skill/cmd/vakio/application/vakio"
	"github.com/konstantin-kukharev/Alice-skill/cmd/vakio/settings"
	"github.com/konstantin-kukharev/Alice-skill/internal/logger"
	"github.com/konstantin-kukharev/Alice-skill/internal/processmanager"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	cfg := settings.New()
	cfg.WithEnv()

	l, err := logger.NewLogger(zap.InfoLevel)
	if err != nil {
		log.Panic(err)
	}
	ctx = l.WithContextFields(ctx,
		zap.Int("pid", os.Getpid()),
		zap.String("app", "vakio client"),
		zap.Any("cfg", cfg))
	defer l.Sync()

	pm := processmanager.NewManager(ctx, 1*time.Second)
	v := vakio.NewTokenApp(
		l,
		cfg.VakioConfig.Login,
		cfg.VakioConfig.CID,
		cfg.VakioConfig.Secret,
		cfg.VakioConfig.Password,
	)
	_ = vakio.NewSourceApp(l, v, cfg.PoolInterval)
	pm.AddTask(v)
	pm.Wait(os.Kill, os.Interrupt)
}
