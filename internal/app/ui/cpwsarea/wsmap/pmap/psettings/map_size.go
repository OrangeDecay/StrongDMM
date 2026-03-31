package psettings

import (
	"fmt"
	"math"

	"sdmm/internal/imguiext"
	"sdmm/internal/imguiext/style"
	w "sdmm/internal/imguiext/widget"

	"github.com/SpaiR/imgui-go"
	"github.com/rs/zerolog/log"
)

const (
	possibleMaxX = math.MaxInt
	possibleMaxY = math.MaxInt
	possibleMaxZ = math.MaxInt
)

type sessionMapSize struct {
	maxX, maxY, maxZ int32
	anchorRight, anchorTop, anchorAbove bool
}

func (s sessionMapSize) String() string {
	return fmt.Sprintf("maxX: %d, maxY: %d, maxZ: %d, anchor: R=%t T=%t A=%t", s.maxX, s.maxY, s.maxZ, s.anchorRight, s.anchorTop, s.anchorAbove)
}

func (p *Panel) DropSessionMapSize() {
	p.sessionMapSize = nil
}

func (p *Panel) showMapSize() {
	if imgui.CollapsingHeader("Map Size") {
		if p.sessionMapSize == nil {
			p.sessionMapSize = &sessionMapSize{
				maxX: int32(p.editor.Dmm().MaxX),
				maxY: int32(p.editor.Dmm().MaxY),
				maxZ: int32(p.editor.Dmm().MaxZ),
			}
		}

		imgui.AlignTextToFramePadding()
		imgui.Text("X")
		imgui.SameLine()
		anchorXText := "Anchor: Left"
		if p.sessionMapSize.anchorRight {
			anchorXText = "Anchor: Right"
		}
		w.Button(anchorXText, func() {
			p.sessionMapSize.anchorRight = !p.sessionMapSize.anchorRight
		}).Size(imgui.Vec2{X: 100}).Build()
		imgui.SameLine()
		imgui.SetNextItemWidth(-1)
		imguiext.InputIntClamp("##max_x", &p.sessionMapSize.maxX, 1, possibleMaxX, 1, 10)

		imgui.AlignTextToFramePadding()
		imgui.Text("Y")
		imgui.SameLine()
		anchorYText := "Anchor: Bottom"
		if p.sessionMapSize.anchorTop {
			anchorYText = "Anchor: Top"
		}
		w.Button(anchorYText, func() {
			p.sessionMapSize.anchorTop = !p.sessionMapSize.anchorTop
		}).Size(imgui.Vec2{X: 100}).Build()
		imgui.SameLine()
		imgui.SetNextItemWidth(-1)
		imguiext.InputIntClamp("##max_y", &p.sessionMapSize.maxY, 1, possibleMaxY, 1, 10)

		imgui.AlignTextToFramePadding()
		imgui.Text("Z")
		imgui.SameLine()
		anchorZText := "Anchor: Under"
		if p.sessionMapSize.anchorAbove {
			anchorZText = "Anchor: Above"
		}
		w.Button(anchorZText, func() {
			p.sessionMapSize.anchorAbove = !p.sessionMapSize.anchorAbove
		}).Size(imgui.Vec2{X: 100}).Build()
		imgui.SameLine()
		imgui.SetNextItemWidth(-1)
		imguiext.InputIntClamp("##max_z", &p.sessionMapSize.maxZ, 1, possibleMaxZ, 1, 10)

		imgui.Separator()

		w.Button("Set", p.doSetMapSize).
			Size(imgui.Vec2{X: -1}).
			Style(style.ButtonGreen{}).
			Build()
	} else {
		p.sessionMapSize = nil
	}
}

func (p *Panel) doSetMapSize() {
	log.Printf("do set map size [%s]: %v", p.editor.Dmm().Name, p.sessionMapSize)
	oldMaxX, oldMaxY, oldMaxZ := p.editor.Dmm().MaxX, p.editor.Dmm().MaxY, p.editor.Dmm().MaxZ
	
	shiftX, shiftY, shiftZ := 0, 0, 0
	if p.sessionMapSize.anchorRight {
		shiftX = int(p.sessionMapSize.maxX) - oldMaxX
	}
	if p.sessionMapSize.anchorTop {
		shiftY = int(p.sessionMapSize.maxY) - oldMaxY
	}
	if p.sessionMapSize.anchorAbove {
		shiftZ = int(p.sessionMapSize.maxZ) - oldMaxZ
	}

	p.editor.Dmm().SetMapSize(int(p.sessionMapSize.maxX), int(p.sessionMapSize.maxY), int(p.sessionMapSize.maxZ), shiftX, shiftY, shiftZ)
	p.editor.CommitMapSizeChange(oldMaxX, oldMaxY, oldMaxZ, shiftX, shiftY, shiftZ)
	p.sessionMapSize = nil
}
