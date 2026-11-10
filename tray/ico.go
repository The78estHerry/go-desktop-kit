package tray

import (
	"bytes"
	"encoding/binary"
	"image/png"
)

// pngToICO wraps PNG data in a single-image ICO container, which Windows
// accepts for tray icons since Vista. Data that is not a PNG is returned
// unchanged, so an ICO passed by the caller still works.
func pngToICO(p []byte) []byte {
	cfg, err := png.DecodeConfig(bytes.NewReader(p))
	if err != nil {
		return p
	}
	dim := func(n int) byte {
		if n >= 256 {
			return 0 // 0 means 256
		}
		return byte(n)
	}
	var buf bytes.Buffer
	w := func(v any) { binary.Write(&buf, binary.LittleEndian, v) }
	w(uint16(0)) // reserved
	w(uint16(1)) // type: icon
	w(uint16(1)) // image count
	buf.WriteByte(dim(cfg.Width))
	buf.WriteByte(dim(cfg.Height))
	buf.WriteByte(0) // palette size
	buf.WriteByte(0) // reserved
	w(uint16(1))     // colour planes
	w(uint16(32))    // bits per pixel
	w(uint32(len(p)))
	w(uint32(6 + 16)) // offset of the image data
	buf.Write(p)
	return buf.Bytes()
}
