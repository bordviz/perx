package canceled

import "context"

func IsContextCanceled(ctx context.Context) error {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}
