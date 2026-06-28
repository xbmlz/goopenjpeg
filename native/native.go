package native

import (
	"os"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	decodeFn func(
		data unsafe.Pointer, dataLen int32, codec int32,
		output *unsafe.Pointer, outputLen *int32,
		width, height, components, precision *int32,
		isSigned *int32,
	) int32
	getParamsFn func(
		data unsafe.Pointer, dataLen int32, codec int32,
		width, height, components, precision *int32,
		isSigned *int32, colourspace *int32,
	) int32
	versionFn func(buf unsafe.Pointer, bufLen int32) int32
	freeFn    func(p unsafe.Pointer)
)

func extractAndLoad(path string) (uintptr, error) {
	if err := os.WriteFile(path, libData, 0o755); err != nil {
		return 0, err
	}
	handle, err := loadLibrary(path)
	if err != nil {
		_ = os.Remove(path)
		return 0, err
	}
	if runtime.GOOS != "windows" {
		_ = os.Remove(path)
	}
	return handle, nil
}

func init() {
	f, err := os.CreateTemp("", "goopenjpeg-*."+libExt())
	if err != nil {
		panic("goopenjpeg: failed to create temp file: " + err.Error())
	}
	path := f.Name()
	_ = f.Close()
	handle, err := extractAndLoad(path)
	if err != nil {
		panic("goopenjpeg: failed to load native library: " + err.Error())
	}
	purego.RegisterLibFunc(&decodeFn, uintptr(handle), "goopenjpeg_decode")
	purego.RegisterLibFunc(&getParamsFn, uintptr(handle), "goopenjpeg_get_parameters")
	purego.RegisterLibFunc(&versionFn, uintptr(handle), "goopenjpeg_version")
	purego.RegisterLibFunc(&freeFn, uintptr(handle), "goopenjpeg_free")
}

func libExt() string {
	switch runtime.GOOS {
	case "windows":
		return "dll"
	case "darwin":
		return "dylib"
	default:
		return "so"
	}
}
