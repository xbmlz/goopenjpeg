#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if ! python3 -c "import ljdata" 2>/dev/null; then
  echo "Installing pylibjpeg-data..."
  pip3 install "git+https://github.com/pydicom/pylibjpeg-data"
fi
if ! python3 -c "import pydicom" 2>/dev/null; then
  echo "Installing pydicom..."
  pip3 install pydicom
fi

python3 - <<'PY'
import json
import pathlib
import shutil

import ljdata
from pydicom.encaps import generate_frames

dst = pathlib.Path("testdata")
dst.mkdir(exist_ok=True)

# ISO 15444 reference codestreams
for rel in (
    "15444/2KLS/693.j2k",
    "15444/2KLS/oj36.j2k",
    "15444/HTJ2K/Bretagne1_ht_lossy.j2k",
    "15444/HTJ2K/Bretagne1_ht.j2k",
):
    src = ljdata.JPEG_DIRECTORY / rel
    target = dst / rel
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, target)
    print(f"copied {src} -> {target}")

REF_DCM = {
    "1.2.840.10008.1.2.4.90": [
        "693_J2KR.dcm",
        "966_fixed.dcm",
        "emri_small_jpeg_2k_lossless.dcm",
        "explicit_VR-UN.dcm",
        "GDCMJ2K_TextGBR.dcm",
        "JPEG2KLossless_1s_1f_u_16_16.dcm",
        "MR_small_jp2klossless.dcm",
        "MR2_J2KR.dcm",
        "NM_Kakadu44_SOTmarkerincons.dcm",
        "RG1_J2KR.dcm",
        "RG3_J2KR.dcm",
        "TOSHIBA_J2K_OpenJPEGv2Regression.dcm",
        "TOSHIBA_J2K_SIZ0_PixRep1.dcm",
        "TOSHIBA_J2K_SIZ1_PixRep0.dcm",
        "US1_J2KR.dcm",
    ],
    "1.2.840.10008.1.2.4.91": [
        "693_J2KI.dcm",
        "ELSCINT1_JP2vsJ2K.dcm",
        "JPEG2000.dcm",
        "MAROTECH_CT_JP2Lossy.dcm",
        "MR2_J2KI.dcm",
        "OsirixFake16BitsStoredFakeSpacing.dcm",
        "RG1_J2KI.dcm",
        "RG3_J2KI.dcm",
        "SC_rgb_gdcm_KY.dcm",
        "US1_J2KI.dcm",
    ],
}

dcm_root = dst / "dcm"
for uid, names in REF_DCM.items():
    index = ljdata.get_indexed_datasets(uid)
    out = dcm_root / uid
    out.mkdir(parents=True, exist_ok=True)
    for fname in names:
        ds = index[fname]["ds"]
        frame = next(generate_frames(ds.PixelData, number_of_frames=1))
        frame_path = out / (fname + ".frame")
        meta_path = out / (fname + ".json")
        frame_path.write_bytes(frame)
        meta = {
            "rows": int(ds.Rows),
            "columns": int(ds.Columns),
            "samples_per_pixel": int(ds.SamplesPerPixel),
            "bits_allocated": int(ds.BitsAllocated),
            "bits_stored": int(getattr(ds, "BitsStored", ds.BitsAllocated)),
            "pixel_representation": int(getattr(ds, "PixelRepresentation", 0)),
            "photometric_interpretation": str(ds.PhotometricInterpretation),
        }
        meta_path.write_text(json.dumps(meta, indent=2))
        print(f"exported {uid}/{fname}")
PY

echo "testdata ready under testdata/15444 and testdata/dcm"
