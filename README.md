# goopenjpeg

Go JPEG 2000 decoder — **no CGO** for callers (`purego` + embedded native library).

Aligned with [pylibjpeg-openjpeg](https://github.com/pydicom/pylibjpeg-openjpeg) for DICOM transfer syntaxes:

| UID | Description |
|-----|-------------|
| 1.2.840.10008.1.2.4.90 | JPEG 2000 Lossless Only |
| 1.2.840.10008.1.2.4.91 | JPEG 2000 |
| 1.2.840.10008.1.2.4.201–203 | HTJ2K |

## Status

**Phase 1 (current):** decode API + CI native builds.

- Done: `DecodeImage`, `GetImageParameters`, `DecodePixelData`, purego loader
- Next: compliance tests, `encode` API (Phase 2)

## Installation

```bash
go get github.com/godicom-dev/goopenjpeg
```

Prebuilt OpenJPEG libraries are embedded per platform in `native/libs/` — no CMake required for `go get` users.

## Usage

### Decode a JPEG 2000 codestream

`stream` may be `[]byte`, a file path (`string`), or `io.Reader`.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/godicom-dev/goopenjpeg"
)

func main() {
	data, err := os.ReadFile("image.j2k")
	if err != nil {
		log.Fatal(err)
	}

	// Shorthand for J2K codestream (0xff 0x4f 0xff 0x51 …)
	img, err := goopenjpeg.Decode(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%dx%d, %d components, precision %d signed=%v\n",
		img.Width, img.Height, img.Components, img.Precision, img.IsSigned)

	// Pixels are planar-interleaved (RGB: R,G,B per pixel), native precision.
	_ = img.Pixels
}
```

JP2 file or explicit codec:

```go
img, err := goopenjpeg.DecodeImage("image.jp2", goopenjpeg.CodecJP2)
```

`Codec` values: `CodecJ2K` (0), `CodecJPT` (1), `CodecJP2` (2).

### Read parameters without decoding pixels

```go
params, err := goopenjpeg.GetParameters(data)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("%dx%d, %d components, precision %d\n",
	params.Width, params.Height, params.Components, params.Precision)
```

### DICOM encapsulated frame

For a single JPEG 2000 frame from DICOM Pixel Data (one item in the encapsulated sequence):

```go
var j2kFrame []byte // one frame from (7FE0,0010)

// Version 2: raw decoded bytes (no extra colour handling)
raw, err := goopenjpeg.DecodePixelData(j2kFrame, goopenjpeg.PixelDataOptions{
	Version: goopenjpeg.PixelDataV2,
	Codec:   goopenjpeg.CodecJ2K,
})

// Version 1: same decode path; PhotometricInterpretation required for API parity
_, err = goopenjpeg.DecodePixelData(j2kFrame, goopenjpeg.PixelDataOptions{
	Version:                   goopenjpeg.PixelDataV1,
	Codec:                     goopenjpeg.CodecJ2K,
	PhotometricInterpretation: "MONOCHROME2",
})
```

### Accessing pixels

```go
// 8-bit sample at (y, x), component c
b := img.ByteAt(y, x, c)

// 16-bit little-endian sample
u := img.Uint16At(y, x, c)
```

### Library version

```go
ver, err := goopenjpeg.OpenJPEGVersion() // e.g. "2.5.4"
```

## API

```go
func DecodeImage(stream any, codec Codec) (*Image, error)
func GetImageParameters(stream any, codec Codec) (*Params, error)
func DecodePixelData(src []byte, opts PixelDataOptions) ([]byte, error)
func OpenJPEGVersion() (string, error)

func Decode(data []byte) (*Image, error)              // CodecJ2K shorthand
func GetParameters(data []byte) (*Params, error)
```

## Platform support

| OS      | amd64 | arm64 |
|---------|-------|-------|
| Windows | ✓     |       |
| macOS   |       | ✓     |
| Linux   | ✓     | ✓     |

## Layout

```
goopenjpeg/           # public Go API
native/               # purego + go:embed prebuilt libs
lib/
  openjpeg/           # submodule → uclouvain/openjpeg
  interface/          # decode glue (from pylibjpeg-openjpeg, memory streams)
  capi/               # C ABI for purego
ref/pylibjpeg-openjpeg/
```

## Development

```bash
git clone --recurse-submodules https://github.com/godicom-dev/goopenjpeg.git
cd goopenjpeg
go test ./...          # uses prebuilt libs in native/libs/
make build-native      # optional: rebuild embedded OpenJPEG (requires CMake)
```

CI (`build.yml`): build-native → commit `native/libs/` on main → test → release on tags.

Tagged releases attach per-platform libraries to GitHub Releases.

## References

- [golibjpeg](https://github.com/godicom-dev/golibjpeg) — same purego architecture for ISO 10918 / JPEG-LS
- [gorle](https://github.com/godicom-dev/gorle) — DICOM RLE Lossless
- [pylibjpeg-openjpeg](https://github.com/pydicom/pylibjpeg-openjpeg) — behaviour and tests reference
