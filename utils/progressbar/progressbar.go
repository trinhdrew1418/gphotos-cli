package progressbar

import (
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"time"
)

func MakeByte(total int64, title string) (*mpb.Progress, *mpb.Bar) {
	p := mpb.New(mpb.WithWidth(64),
		mpb.WithRefreshRate(180*time.Millisecond))
	bar := p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(title),
			decor.CountersKibiByte("% 6.1f / % 6.1f"),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
			decor.Name(" ] [ "),
			decor.EwmaETA(decor.ET_STYLE_MMSS, float64(total)/2048),
			decor.Name(" ] [ "),
			decor.AverageSpeed(decor.UnitKiB, "% .2f"),
		),
	)

	return p, bar
}

func MakeCount(total int64, title string) (*mpb.Progress, *mpb.Bar) {
	p := mpb.New(mpb.WithWidth(64))

	bar := p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(title),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),

		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)

	return p, bar
}
