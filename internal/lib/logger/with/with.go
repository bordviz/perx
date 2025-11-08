package with

import "log/slog"

func WithOp(log *slog.Logger, op string) *slog.Logger {
	return log.With(
		slog.String("op", op),
	)
}
