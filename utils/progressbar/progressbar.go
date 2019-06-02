package progressbar

import (
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func Make(total int64) *mpb.Bar {
	p := mpb.New(mpb.WithWidth(64))
	bar := p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name("Files: "),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
	)
	return bar
}
