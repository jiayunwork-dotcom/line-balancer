package line

import "context"

func abortFresh() error {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ctx.Err()
	if err != nil {
		return err
	}
	return nil
}
