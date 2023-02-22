package src

import "go.uber.org/zap"

func Start(args *Args, logger *zap.SugaredLogger) {
	logger.Infof("Starting download loop for %s", args.ParseRoot)
}
