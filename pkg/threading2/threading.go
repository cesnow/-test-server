package threading2

import (
	"context"

	"github.com/zeromicro/go-zero/core/contextx"
	"github.com/zeromicro/go-zero/core/rescue"
)

func GoSafeContext(ctx context.Context, goF func(ctx context.Context)) {
	if goF != nil {
		ctx = contextx.ValueOnlyFrom(ctx)
		go func(ctx context.Context) {
			defer rescue.Recover()
			goF(ctx)
		}(ctx)
	}
}
