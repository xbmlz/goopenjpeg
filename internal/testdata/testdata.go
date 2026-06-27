package testdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Root() string {
	return filepath.Join("testdata")
}

func JPEGPath(parts ...string) string {
	all := append([]string{Root()}, parts...)
	return filepath.Join(all...)
}

func RequireFile(t *testing.T, parts ...string) []byte {
	t.Helper()
	path := JPEGPath(parts...)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("reference testdata not installed at %s (run scripts/fetch-testdata.sh)", path)
	}
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type DCMFrameMeta struct {
	Rows                      int    `json:"rows"`
	Columns                   int    `json:"columns"`
	SamplesPerPixel           int    `json:"samples_per_pixel"`
	BitsAllocated             int    `json:"bits_allocated"`
	BitsStored                int    `json:"bits_stored"`
	PixelRepresentation       int    `json:"pixel_representation"`
	PhotometricInterpretation string `json:"photometric_interpretation"`
}

func (m DCMFrameMeta) IsSigned() bool {
	return m.PixelRepresentation == 1
}

func DCMPath(uid, name string) string {
	return filepath.Join(Root(), "dcm", uid, name)
}

func RequireDCMFrame(t *testing.T, uid, dcmName string) ([]byte, DCMFrameMeta) {
	t.Helper()
	base := DCMPath(uid, dcmName)
	framePath := base + ".frame"
	metaPath := base + ".json"

	frame, err := os.ReadFile(framePath)
	if os.IsNotExist(err) {
		t.Skipf("DICOM testdata not installed at %s (run scripts/fetch-testdata.sh)", framePath)
	}
	if err != nil {
		t.Fatal(err)
	}

	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatal(err)
	}
	var meta DCMFrameMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		t.Fatal(err)
	}
	return frame, meta
}
