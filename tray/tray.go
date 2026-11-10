// Package tray shows an icon with a menu in the system tray (Windows
// notification area, macOS menu bar, Linux StatusNotifierItem).
//
// It wraps fyne.io/systray behind callbacks and plain PNG icons, and takes
// care of the platform details: on Windows the tray runs on its own locked
// OS thread so a GUI toolkit such as Gio can keep the main goroutine, and
// PNG icons are converted to ICO.
//
//	tray.Start(func(t *tray.Tray) {
//		t.SetIcon(iconPNG)
//		t.SetTooltip("My App")
//		t.AddItem("Show window", showWindow)
//		auto := t.AddCheckbox("Start at sign-in", false, nil)
//		auto.OnClick(func() { auto.SetChecked(!auto.Checked()) })
//		t.AddSeparator()
//		t.AddItem("Exit", t.Quit)
//	}, func() { os.Exit(0) })
//
// Callbacks run on a goroutine owned by the package, never on the GUI
// thread; hand work to the GUI with its own mechanism (for Gio,
// Window.Invalidate plus a channel).
package tray

import (
	"sync"

	"fyne.io/systray"
)

// Tray is the tray icon and its top-level menu.
type Tray struct{}

// Start shows the tray icon and returns at once. ready builds the menu;
// onExit, which may be nil, runs after Quit.
func Start(ready func(*Tray), onExit func()) {
	t := &Tray{}
	start(func() { ready(t) }, onExit)
}

// SetIcon sets the icon from PNG data; a square image of 32 or 64 pixels
// works everywhere.
func (t *Tray) SetIcon(png []byte) { systray.SetIcon(platformIcon(png)) }

// SetTooltip sets the text shown when the pointer rests on the icon.
func (t *Tray) SetTooltip(s string) { systray.SetTooltip(s) }

// SetTitle sets the title: text next to the icon on macOS, the
// application name on Linux. Windows ignores it.
func (t *Tray) SetTitle(s string) { systray.SetTitle(s) }

// OnTapped sets what a primary click on the icon does; the menu then opens
// with a secondary click. Without it a primary click opens the menu.
func (t *Tray) OnTapped(f func()) { systray.SetOnTapped(f) }

// AddItem adds a menu item; onClick may be nil.
func (t *Tray) AddItem(title string, onClick func()) *Item {
	return newItem(systray.AddMenuItem(title, ""), onClick)
}

// AddCheckbox adds an item with a check mark. Clicking does not toggle it;
// call SetChecked from onClick once the action has succeeded.
func (t *Tray) AddCheckbox(title string, checked bool, onClick func()) *Item {
	return newItem(systray.AddMenuItemCheckbox(title, "", checked), onClick)
}

// AddSeparator adds a line between items.
func (t *Tray) AddSeparator() { systray.AddSeparator() }

// Quit removes the icon; onExit passed to Start runs afterwards.
func (t *Tray) Quit() { systray.Quit() }

// Item is one menu entry.
type Item struct {
	mi      *systray.MenuItem
	mu      sync.Mutex
	onClick func()
}

func newItem(mi *systray.MenuItem, onClick func()) *Item {
	it := &Item{mi: mi, onClick: onClick}
	go func() {
		for range mi.ClickedCh {
			it.mu.Lock()
			f := it.onClick
			it.mu.Unlock()
			if f != nil {
				f()
			}
		}
	}()
	return it
}

// OnClick replaces what clicking the item does.
func (it *Item) OnClick(f func()) {
	it.mu.Lock()
	it.onClick = f
	it.mu.Unlock()
}

// AddItem adds an item to this item's submenu.
func (it *Item) AddItem(title string, onClick func()) *Item {
	return newItem(it.mi.AddSubMenuItem(title, ""), onClick)
}

// AddCheckbox adds a checkbox item to this item's submenu.
func (it *Item) AddCheckbox(title string, checked bool, onClick func()) *Item {
	return newItem(it.mi.AddSubMenuItemCheckbox(title, "", checked), onClick)
}

// AddSeparator adds a line to this item's submenu.
func (it *Item) AddSeparator() { it.mi.AddSeparator() }

// SetTitle changes the text.
func (it *Item) SetTitle(s string) { it.mi.SetTitle(s) }

// SetTooltip sets the text shown when the pointer rests on the item, where
// the platform supports it.
func (it *Item) SetTooltip(s string) { it.mi.SetTooltip(s) }

// SetEnabled enables or greys out the item.
func (it *Item) SetEnabled(on bool) {
	if on {
		it.mi.Enable()
	} else {
		it.mi.Disable()
	}
}

// SetVisible shows or hides the item.
func (it *Item) SetVisible(on bool) {
	if on {
		it.mi.Show()
	} else {
		it.mi.Hide()
	}
}

// SetChecked sets the check mark.
func (it *Item) SetChecked(on bool) {
	if on {
		it.mi.Check()
	} else {
		it.mi.Uncheck()
	}
}

// Checked reports whether the item has a check mark.
func (it *Item) Checked() bool { return it.mi.Checked() }

// SetIcon sets a small icon in front of the text from PNG data.
func (it *Item) SetIcon(png []byte) { it.mi.SetIcon(platformIcon(png)) }

// Remove deletes the item from the menu.
func (it *Item) Remove() { it.mi.Remove() }
