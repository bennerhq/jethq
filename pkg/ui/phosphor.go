package ui

import (
	"bytes"
	_ "embed"
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const phosphorRasterSize = 128

// Phosphor Regular toolbar assets. The source SVGs are bundled so their
// monochrome currentColor paths can be rendered in every active UI color.
//
//go:embed assets/icons/phosphor-regular/reconnect.svg
var phosphorReconnectSVG []byte

//go:embed assets/icons/phosphor-regular/paste.svg
var phosphorPasteSVG []byte

//go:embed assets/icons/phosphor-regular/media.svg
var phosphorMediaSVG []byte

//go:embed assets/icons/phosphor-regular/stats.svg
var phosphorStatsSVG []byte

//go:embed assets/icons/phosphor-regular/terminal.svg
var phosphorTerminalSVG []byte

//go:embed assets/icons/phosphor-regular/fullscreen.svg
var phosphorFullscreenSVG []byte

//go:embed assets/icons/phosphor-regular/settings.svg
var phosphorSettingsSVG []byte

var (
	phosphorIconsOnce sync.Once
	phosphorIcons     map[IconKind]*ebiten.Image
)

func phosphorIcon(kind IconKind) *ebiten.Image {
	phosphorIconsOnce.Do(func() {
		phosphorIcons = make(map[IconKind]*ebiten.Image)
		for kind, svg := range map[IconKind][]byte{
			IconReconnect:  phosphorReconnectSVG,
			IconPaste:      phosphorPasteSVG,
			IconMedia:      phosphorMediaSVG,
			IconStats:      phosphorStatsSVG,
			IconTerminal:   phosphorTerminalSVG,
			IconFullscreen: phosphorFullscreenSVG,
			IconSettings:   phosphorSettingsSVG,
		} {
			if image := renderPhosphorIcon(svg); image != nil {
				phosphorIcons[kind] = image
			}
		}
	})
	return phosphorIcons[kind]
}

func renderPhosphorIcon(svg []byte) *ebiten.Image {
	icon, err := oksvg.ReadReplacingCurrentColor(bytes.NewReader(svg), "#ffffff")
	if err != nil {
		return nil
	}
	icon.SetTarget(0, 0, phosphorRasterSize, phosphorRasterSize)
	raster := image.NewRGBA(image.Rect(0, 0, phosphorRasterSize, phosphorRasterSize))
	scanner := rasterx.NewScannerGV(phosphorRasterSize, phosphorRasterSize, raster, raster.Bounds())
	icon.Draw(rasterx.NewDasher(phosphorRasterSize, phosphorRasterSize, scanner), 1)
	return ebiten.NewImageFromImage(raster)
}
