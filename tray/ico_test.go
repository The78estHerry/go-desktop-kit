package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"testing"
)

func TestPNGToICO(t *testing.T) {
	for _, size := range []int{32, 256} {
		var p bytes.Buffer
		if err := png.Encode(&p, image.NewNRGBA(image.Rect(0, 0, size, size))); err != nil {
			t.Fatal(err)
		}
		ico := pngToICO(p.Bytes())
		if len(ico) != 22+p.Len() {
			t.Fatalf("size %d: length %d, want %d", size, len(ico), 22+p.Len())
		}
		if typ := binary.LittleEndian.Uint16(ico[2:]); typ != 1 {
			t.Errorf("size %d: type %d, want 1", size, typ)
		}
		want := byte(size)
		if size == 256 {
			want = 0
		}
		if ico[6] != want || ico[7] != want {
			t.Errorf("size %d: dimensions %d×%d, want %d", size, ico[6], ico[7], want)
		}
		if n := binary.LittleEndian.Uint32(ico[14:]); int(n) != p.Len() {
			t.Errorf("size %d: data length %d", size, n)
		}
		if !bytes.Equal(ico[22:], p.Bytes()) {
			t.Errorf("size %d: PNG data changed", size)
		}
	}
	if got := pngToICO([]byte("not a png")); string(got) != "not a png" {
		t.Errorf("non-PNG data was changed: %q", got)
	}
}
