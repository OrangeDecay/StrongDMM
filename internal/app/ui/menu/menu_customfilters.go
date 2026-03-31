package menu

import (
	"fmt"
	"strings"

	w "sdmm/internal/imguiext/widget"

	"github.com/SpaiR/imgui-go"
)

const addFilterPopupID = "Add Custom Filter###add_custom_filter_popup"

// showCustomFilters renders the custom filters section inside the View menu.
// Each filter gets a selectable row with a [×] delete button.
// At the bottom an "Add Custom Filter…" item opens the popup.
func (m *Menu) showCustomFilters() {
	filters := m.app.PathsFilter().CustomFilters

	for i, f := range filters {
		// Selectable acts as the toggle checkbox.
		selected := !f.Hidden
		label := fmt.Sprintf("%s##custom_filter_%d", f.Name, i)
		if imgui.SelectableV(label, selected, imgui.SelectableFlagsNone, imgui.Vec2{}) {
			m.app.DoToggleCustomFilter(i)
		}
		if imgui.IsItemHovered() {
			// Show the paths this filter covers as a tooltip.
			imgui.BeginTooltip()
			imgui.Text("Paths:\n" + strings.Join(f.Paths, "\n"))
			imgui.EndTooltip()
		}
		imgui.SameLine()
		// [×] delete button — placed at the right margin.
		delLabel := fmt.Sprintf("%s##del_custom_filter_%d", "×", i)
		if imgui.SmallButton(delLabel) {
			m.app.DoRemoveCustomFilter(i)
			break // slice changed; safe to break and re-render next frame
		}
	}

	w.Separator().Build()
	if imgui.MenuItemV("Add Custom Filter...", "", false, true) {
		m.openAddFilterPopup()
	}

	// Render the popup modal (must be called every frame).
	m.processAddFilterPopup()
}

func (m *Menu) openAddFilterPopup() {
	m.addFilterOpen = true
	m.addFilterName = ""
	m.addFilterPathInput = ""
	m.addFilterPaths = nil
	imgui.OpenPopup(addFilterPopupID)
}

func (m *Menu) processAddFilterPopup() {
	if !m.addFilterOpen {
		return
	}

	imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 0})
	if imgui.BeginPopupModalV(addFilterPopupID, nil, 0) {
		imgui.Text("Filter name:")
		imgui.SetNextItemWidth(-1)
		imgui.InputTextV("##add_filter_name", &m.addFilterName, 0, nil)

		imgui.Spacing()
		imgui.Text("Add type path (e.g. /obj/item/weapon):")
		imgui.SetNextItemWidth(-60)
		imgui.InputTextV("##add_filter_path_input", &m.addFilterPathInput, 0, nil)
		imgui.SameLine()
		if imgui.ButtonV("Add##add_filter_path_add", imgui.Vec2{}) && strings.TrimSpace(m.addFilterPathInput) != "" {
			path := strings.TrimSpace(m.addFilterPathInput)
			m.addFilterPaths = append(m.addFilterPaths, path)
			m.addFilterPathInput = ""
		}

		if len(m.addFilterPaths) > 0 {
			imgui.Spacing()
			imgui.Text("Paths in filter:")
			for i, p := range m.addFilterPaths {
				imgui.BulletText(p)
				imgui.SameLine()
				if imgui.SmallButton(fmt.Sprintf("×##del_new_path_%d", i)) {
					m.addFilterPaths = append(m.addFilterPaths[:i], m.addFilterPaths[i+1:]...)
					break
				}
			}
		}

		imgui.Spacing()
		imgui.Separator()

		canCreate := strings.TrimSpace(m.addFilterName) != "" && len(m.addFilterPaths) > 0
		if !canCreate {
			imgui.BeginDisabledV(true)
		}
		if imgui.ButtonV("Create##add_filter_create", imgui.Vec2{}) {
			m.app.DoAddCustomFilter(strings.TrimSpace(m.addFilterName), m.addFilterPaths)
			m.addFilterOpen = false
			imgui.CloseCurrentPopup()
		}
		if !canCreate {
			imgui.EndDisabled()
		}
		imgui.SameLine()
		if imgui.ButtonV("Cancel##add_filter_cancel", imgui.Vec2{}) {
			m.addFilterOpen = false
			imgui.CloseCurrentPopup()
		}

		imgui.EndPopup()
	}
}
