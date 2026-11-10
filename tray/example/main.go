// Command example shows a tray icon and doubles as a manual test: a
// primary click changes the colour, the menu has a disabled status line, a
// checkbox, a submenu, a hidden item and Exit.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/The78estHerry/go-desktop-kit/tray"
)

var colors = []struct {
	name string
	c    color.NRGBA
}{
	{"Green", color.NRGBA{0x2e, 0xa0, 0x43, 0xff}},
	{"Blue", color.NRGBA{0x1f, 0x6f, 0xeb, 0xff}},
	{"Red", color.NRGBA{0xd1, 0x24, 0x2f, 0xff}},
}

// circle draws a filled circle on a transparent square.
func circle(c color.NRGBA, size int) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	r := float64(size)/2 - 1
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)+0.5-float64(size)/2, float64(y)+0.5-float64(size)/2
			if dx*dx+dy*dy <= r*r {
				img.SetNRGBA(x, y, c)
			}
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}

func main() {
	tray.Start(func(t *tray.Tray) {
		cur := 0
		status := t.AddItem("", nil)
		status.SetEnabled(false)
		apply := func(i int) {
			cur = i
			t.SetIcon(circle(colors[i].c, 64))
			t.SetTooltip("tray example: " + colors[i].name)
			status.SetTitle("Colour: " + colors[i].name)
		}
		apply(0)
		t.OnTapped(func() { apply((cur + 1) % len(colors)) })

		t.AddSeparator()
		sub := t.AddItem("Colour", nil)
		for i, c := range colors {
			item := sub.AddItem(c.name, func() { apply(i) })
			item.SetIcon(circle(c.c, 16))
		}
		check := t.AddCheckbox("Checkbox (click to toggle)", false, nil)
		check.OnClick(func() { check.SetChecked(!check.Checked()) })
		hidden := t.AddItem("Hidden item: you found it", nil)
		hidden.SetVisible(false)
		show := t.AddCheckbox("Show hidden item", false, nil)
		show.OnClick(func() {
			on := !show.Checked()
			show.SetChecked(on)
			hidden.SetVisible(on)
		})
		t.AddSeparator()
		t.AddItem("Exit", t.Quit)
	}, func() {
		fmt.Println("tray closed")
		os.Exit(0)
	})
	select {} // the tray runs on its own goroutine
}
