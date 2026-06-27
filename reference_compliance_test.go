package goopenjpeg_test

import (
	"encoding/binary"
	"testing"

	"github.com/godicom-dev/goopenjpeg"
	"github.com/godicom-dev/goopenjpeg/internal/testdata"
)

func int16At(img *goopenjpeg.Image, y, x, c int) int16 {
	bps := img.BytesPerSample()
	off := (y*img.Width+x)*img.Components*bps + c*bps
	return int16(binary.LittleEndian.Uint16(img.Pixels[off:]))
}

func rgbAt(img *goopenjpeg.Image, y, x int) [3]int {
	var px [3]int
	for c := 0; c < 3; c++ {
		px[c] = int(img.ByteAt(y, x, c))
	}
	return px
}

func detectCodec(data []byte) goopenjpeg.Codec {
	if len(data) >= 8 && data[4] == 'j' && data[5] == 'P' && data[6] == ' ' && data[7] == ' ' {
		return goopenjpeg.CodecJP2
	}
	return goopenjpeg.CodecJ2K
}

func decodeDCMFrame(t *testing.T, frame []byte) (*goopenjpeg.Image, error) {
	t.Helper()
	return goopenjpeg.DecodeImage(frame, detectCodec(frame))
}

func paramsDCMFrame(t *testing.T, frame []byte) (*goopenjpeg.Params, error) {
	t.Helper()
	return goopenjpeg.GetImageParameters(frame, detectCodec(frame))
}

func TestReferenceJ2KLS693(t *testing.T) {
	data := testdata.RequireFile(t, "15444", "2KLS", "693.j2k")
	img, err := goopenjpeg.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{340, 815, 1229, 1358, 1351, 1302, 1069, 618, 215, 71}
	for i, x := range []int{55, 56, 57, 58, 59, 60, 61, 62, 63, 64} {
		if got := int16At(img, 270, x, 0); got != int16(want[i]) {
			t.Fatalf("arr[270,%d]: got %d want %d", x, got, want[i])
		}
	}
}

func TestReferenceJ2KLSOj36Decode(t *testing.T) {
	data := testdata.RequireFile(t, "15444", "2KLS", "oj36.j2k")
	img, err := goopenjpeg.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if img.Width != 256 || img.Height != 256 || img.Components != 3 {
		t.Fatalf("unexpected image: %+v", img)
	}
	if got := rgbAt(img, 0, 0); got != [3]int{235, 244, 245} {
		t.Fatalf("arr[0,0]: got %v want [235 244 245]", got)
	}
	want := [][3]int{
		{160, 171, 199}, {174, 182, 193}, {190, 198, 209}, {209, 217, 213},
		{219, 227, 223}, {226, 235, 221}, {233, 242, 228}, {239, 246, 236},
		{243, 250, 240}, {247, 250, 248},
	}
	for i, x := range []int{35, 36, 37, 38, 39, 40, 41, 42, 43, 44} {
		if got := rgbAt(img, 60, x); got != want[i] {
			t.Fatalf("arr[60,%d]: got %v want %v", x, got, want[i])
		}
	}
}

func TestReferenceJ2KLSOj36Parameters(t *testing.T) {
	data := testdata.RequireFile(t, "15444", "2KLS", "oj36.j2k")
	p, err := goopenjpeg.GetParameters(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Height != 256 || p.Width != 256 || p.Components != 3 || p.Precision != 8 || p.IsSigned {
		t.Fatalf("unexpected params: %+v", p)
	}
}

func TestReferenceHTJ2KLossy(t *testing.T) {
	data := testdata.RequireFile(t, "15444", "HTJ2K", "Bretagne1_ht_lossy.j2k")
	img, err := goopenjpeg.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	wantRow := [][3]int{
		{91, 37, 2}, {94, 40, 1}, {97, 42, 5}, {174, 123, 59}, {172, 132, 69},
		{169, 134, 74}, {168, 136, 77}, {168, 137, 80}, {168, 136, 80}, {169, 136, 78},
	}
	for i, x := range []int{295, 296, 297, 298, 299, 300, 301, 302, 303, 304} {
		if got := rgbAt(img, 160, x); got != wantRow[i] {
			t.Fatalf("arr[160,%d]: got %v want %v", x, got, wantRow[i])
		}
	}
}

func TestReferenceHTJ2KLossless(t *testing.T) {
	data := testdata.RequireFile(t, "15444", "HTJ2K", "Bretagne1_ht.j2k")
	img, err := goopenjpeg.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := rgbAt(img, 160, 295); got != [3]int{90, 38, 1} {
		t.Fatalf("arr[160,295]: got %v", got)
	}
}

type refDCMCase struct {
	uid    string
	name   string
	rows   int
	cols   int
	spp    int
	bps    int
	signed bool
}

// decode overrides when JPEG codestream metadata differs from pylibjpeg decode tables.
var refDCMDecodeBPS = map[string]int{
	"OsirixFake16BitsStoredFakeSpacing.dcm": 11,
}

var refDCMDecodeSigned = map[string]bool{
	"TOSHIBA_J2K_OpenJPEGv2Regression.dcm": false,
	"TOSHIBA_J2K_SIZ0_PixRep1.dcm":         false,
	"TOSHIBA_J2K_SIZ1_PixRep0.dcm":         true,
}

type refDCMParamsCase struct {
	uid      string
	name     string
	rows     int
	cols     int
	spp      int
	bps      int
	signed   bool
	paramBPS int // when different from decode bps (e.g. OsirixFake)
}

var refDCM90 = []refDCMCase{
	{"1.2.840.10008.1.2.4.90", "693_J2KR.dcm", 512, 512, 1, 14, true},
	{"1.2.840.10008.1.2.4.90", "966_fixed.dcm", 2128, 2000, 1, 12, false},
	{"1.2.840.10008.1.2.4.90", "emri_small_jpeg_2k_lossless.dcm", 64, 64, 1, 16, false},
	{"1.2.840.10008.1.2.4.90", "explicit_VR-UN.dcm", 512, 512, 1, 16, true},
	{"1.2.840.10008.1.2.4.90", "GDCMJ2K_TextGBR.dcm", 400, 400, 3, 8, false},
	{"1.2.840.10008.1.2.4.90", "JPEG2KLossless_1s_1f_u_16_16.dcm", 1416, 1420, 1, 16, false},
	{"1.2.840.10008.1.2.4.90", "MR_small_jp2klossless.dcm", 64, 64, 1, 16, true},
	{"1.2.840.10008.1.2.4.90", "MR2_J2KR.dcm", 1024, 1024, 1, 12, false},
	{"1.2.840.10008.1.2.4.90", "NM_Kakadu44_SOTmarkerincons.dcm", 2500, 2048, 1, 16, false},
	{"1.2.840.10008.1.2.4.90", "RG1_J2KR.dcm", 1955, 1841, 1, 15, false},
	{"1.2.840.10008.1.2.4.90", "RG3_J2KR.dcm", 1760, 1760, 1, 10, false},
	{"1.2.840.10008.1.2.4.90", "TOSHIBA_J2K_OpenJPEGv2Regression.dcm", 512, 512, 1, 16, true},
	{"1.2.840.10008.1.2.4.90", "TOSHIBA_J2K_SIZ0_PixRep1.dcm", 512, 512, 1, 16, true},
	{"1.2.840.10008.1.2.4.90", "TOSHIBA_J2K_SIZ1_PixRep0.dcm", 512, 512, 1, 16, false},
	{"1.2.840.10008.1.2.4.90", "US1_J2KR.dcm", 480, 640, 3, 8, false},
}

var refDCM91 = []refDCMCase{
	{"1.2.840.10008.1.2.4.91", "693_J2KI.dcm", 512, 512, 1, 16, true},
	{"1.2.840.10008.1.2.4.91", "ELSCINT1_JP2vsJ2K.dcm", 512, 512, 1, 12, false},
	{"1.2.840.10008.1.2.4.91", "JPEG2000.dcm", 1024, 256, 1, 16, true},
	{"1.2.840.10008.1.2.4.91", "MAROTECH_CT_JP2Lossy.dcm", 716, 512, 1, 12, false},
	{"1.2.840.10008.1.2.4.91", "MR2_J2KI.dcm", 1024, 1024, 1, 12, false},
	{"1.2.840.10008.1.2.4.91", "OsirixFake16BitsStoredFakeSpacing.dcm", 224, 176, 1, 16, false},
	{"1.2.840.10008.1.2.4.91", "RG1_J2KI.dcm", 1955, 1841, 1, 15, false},
	{"1.2.840.10008.1.2.4.91", "RG3_J2KI.dcm", 1760, 1760, 1, 10, false},
	{"1.2.840.10008.1.2.4.91", "SC_rgb_gdcm_KY.dcm", 100, 100, 3, 8, false},
	{"1.2.840.10008.1.2.4.91", "US1_J2KI.dcm", 480, 640, 3, 8, false},
}

// Parameter tests follow pylibjpeg-openjpeg test_parameters.py (may differ from decode on signed/bps).
var refDCM90Params = []refDCMParamsCase{
	{"1.2.840.10008.1.2.4.90", "693_J2KR.dcm", 512, 512, 1, 14, true, 0},
	{"1.2.840.10008.1.2.4.90", "966_fixed.dcm", 2128, 2000, 1, 12, false, 0},
	{"1.2.840.10008.1.2.4.90", "emri_small_jpeg_2k_lossless.dcm", 64, 64, 1, 16, false, 0},
	{"1.2.840.10008.1.2.4.90", "explicit_VR-UN.dcm", 512, 512, 1, 16, true, 0},
	{"1.2.840.10008.1.2.4.90", "GDCMJ2K_TextGBR.dcm", 400, 400, 3, 8, false, 0},
	{"1.2.840.10008.1.2.4.90", "JPEG2KLossless_1s_1f_u_16_16.dcm", 1416, 1420, 1, 16, false, 0},
	{"1.2.840.10008.1.2.4.90", "MR_small_jp2klossless.dcm", 64, 64, 1, 16, true, 0},
	{"1.2.840.10008.1.2.4.90", "MR2_J2KR.dcm", 1024, 1024, 1, 12, false, 0},
	{"1.2.840.10008.1.2.4.90", "NM_Kakadu44_SOTmarkerincons.dcm", 2500, 2048, 1, 16, false, 0},
	{"1.2.840.10008.1.2.4.90", "RG1_J2KR.dcm", 1955, 1841, 1, 15, false, 0},
	{"1.2.840.10008.1.2.4.90", "RG3_J2KR.dcm", 1760, 1760, 1, 10, false, 0},
	{"1.2.840.10008.1.2.4.90", "TOSHIBA_J2K_OpenJPEGv2Regression.dcm", 512, 512, 1, 16, false, 0},
	{"1.2.840.10008.1.2.4.90", "TOSHIBA_J2K_SIZ0_PixRep1.dcm", 512, 512, 1, 16, false, 0},
	{"1.2.840.10008.1.2.4.90", "TOSHIBA_J2K_SIZ1_PixRep0.dcm", 512, 512, 1, 16, true, 0},
	{"1.2.840.10008.1.2.4.90", "US1_J2KR.dcm", 480, 640, 3, 8, false, 0},
}

var refDCM91Params = []refDCMParamsCase{
	{"1.2.840.10008.1.2.4.91", "693_J2KI.dcm", 512, 512, 1, 16, true, 0},
	{"1.2.840.10008.1.2.4.91", "ELSCINT1_JP2vsJ2K.dcm", 512, 512, 1, 12, false, 0},
	{"1.2.840.10008.1.2.4.91", "JPEG2000.dcm", 1024, 256, 1, 16, true, 0},
	{"1.2.840.10008.1.2.4.91", "MAROTECH_CT_JP2Lossy.dcm", 716, 512, 1, 12, false, 0},
	{"1.2.840.10008.1.2.4.91", "MR2_J2KI.dcm", 1024, 1024, 1, 12, false, 0},
	{"1.2.840.10008.1.2.4.91", "OsirixFake16BitsStoredFakeSpacing.dcm", 224, 176, 1, 16, false, 11},
	{"1.2.840.10008.1.2.4.91", "RG1_J2KI.dcm", 1955, 1841, 1, 15, false, 0},
	{"1.2.840.10008.1.2.4.91", "RG3_J2KI.dcm", 1760, 1760, 1, 10, false, 0},
	{"1.2.840.10008.1.2.4.91", "SC_rgb_gdcm_KY.dcm", 100, 100, 3, 8, false, 0},
	{"1.2.840.10008.1.2.4.91", "US1_J2KI.dcm", 480, 640, 3, 8, false, 0},
}

func paramPrecision(tc refDCMParamsCase) int {
	if tc.paramBPS > 0 {
		return tc.paramBPS
	}
	return tc.bps
}

func decodePrecision(tc refDCMCase) int {
	if bps, ok := refDCMDecodeBPS[tc.name]; ok {
		return bps
	}
	return tc.bps
}

func decodeSigned(tc refDCMCase) bool {
	if signed, ok := refDCMDecodeSigned[tc.name]; ok {
		return signed
	}
	return tc.signed
}

func TestReferenceDCMGetParameters(t *testing.T) {
	cases := append(append([]refDCMParamsCase{}, refDCM90Params...), refDCM91Params...)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frame, _ := testdata.RequireDCMFrame(t, tc.uid, tc.name)
			p, err := paramsDCMFrame(t, frame)
			if err != nil {
				t.Fatal(err)
			}
			wantBPS := paramPrecision(tc)
			if p.Height != tc.rows || p.Width != tc.cols {
				t.Fatalf("dims: got %dx%d want %dx%d", p.Height, p.Width, tc.rows, tc.cols)
			}
			if p.Components != tc.spp || p.Precision != wantBPS {
				t.Fatalf("components/precision: got %d/%d want %d/%d", p.Components, p.Precision, tc.spp, wantBPS)
			}
			if p.IsSigned != tc.signed {
				t.Fatalf("signed: got %v want %v", p.IsSigned, tc.signed)
			}
		})
	}
}

func TestReferenceDCMDecode(t *testing.T) {
	cases := append(append([]refDCMCase{}, refDCM90...), refDCM91...)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frame, _ := testdata.RequireDCMFrame(t, tc.uid, tc.name)
			img, err := decodeDCMFrame(t, frame)
			if err != nil {
				t.Fatal(err)
			}
			wantBPS := decodePrecision(tc)
			wantSigned := decodeSigned(tc)
			if img.Height != tc.rows || img.Width != tc.cols {
				t.Fatalf("dims: got %dx%d want %dx%d", img.Height, img.Width, tc.rows, tc.cols)
			}
			if img.Components != tc.spp || img.Precision != wantBPS {
				t.Fatalf("components/precision: got %d/%d want %d/%d", img.Components, img.Precision, tc.spp, wantBPS)
			}
			if img.IsSigned != wantSigned {
				t.Fatalf("signed: got %v want %v", img.IsSigned, wantSigned)
			}
		})
	}
}

func TestReferenceDCMHandlerPixels(t *testing.T) {
	uid := "1.2.840.10008.1.2.4.90"
	frame, _ := testdata.RequireDCMFrame(t, uid, "693_J2KR.dcm")
	img, err := decodeDCMFrame(t, frame)
	if err != nil {
		t.Fatal(err)
	}
	if got := int16At(img, 0, 0, 0); got != -2000 {
		t.Fatalf("arr[0,0]: got %d want -2000", got)
	}

	frame, _ = testdata.RequireDCMFrame(t, uid, "MR_small_jp2klossless.dcm")
	img, err = decodeDCMFrame(t, frame)
	if err != nil {
		t.Fatal(err)
	}
	if got := int16At(img, 0, 31, 0); got != 422 {
		t.Fatalf("arr[0,31]: got %d want 422", got)
	}
	if got := int16At(img, 31, 0, 0); got != 366 {
		t.Fatalf("arr[31,0]: got %d want 366", got)
	}
}
