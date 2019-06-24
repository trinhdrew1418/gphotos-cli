package progressbar

import (
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func Make(total int64, title string) (*mpb.Progress, *mpb.Bar) {
	p := mpb.New(mpb.WithWidth(64))
	bar := p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(title),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
	)

	return p, bar
}
