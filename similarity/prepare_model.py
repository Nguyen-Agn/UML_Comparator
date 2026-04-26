"""
prepare_model.py — Convert paraphrase-multilingual-MiniLM-L12-v2 to ONNX (INT8 quantized)
and package it into minilm.zip for the Go semantic matcher.

Usage:
    python3 AI_translate/prepare_model.py

Output:
    AI_translate/minilm.zip  (contains model.onnx, tokenizer.json, onnxruntime.dll)

Requirements:
    pip install transformers optimum[onnxruntime] onnxruntime
"""

import os
import sys

# Force D:\pylibs to be the ONLY source for ML packages.
# Remove any pre-loaded conflicting packages from the system Python.
_PYLIBS = r"D:/pylibs"

# Step 1: Insert pylibs at the very front
if _PYLIBS in sys.path:
    sys.path.remove(_PYLIBS)
sys.path.insert(0, _PYLIBS)

# Step 2: Purge any already-imported versions of the ML stack from module cache
_purge_prefixes = (
    "transformers", "huggingface_hub", "optimum",
    "tokenizers", "safetensors", "onnx", "onnxruntime",
)
for key in list(sys.modules.keys()):
    if any(key == p or key.startswith(p + ".") for p in _purge_prefixes):
        del sys.modules[key]

import shutil
import zipfile
import urllib.request
import tempfile

MODEL_NAME = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
OUTPUT_ZIP = os.path.join(SCRIPT_DIR, "minilm.zip")

# ONNX Runtime DLL — must match onnxruntime_go v1.28.0 which requires ORT v1.25.0 (API v25)
ORT_VERSION = "1.25.0"
ORT_NUGET_URL = f"https://www.nuget.org/api/v2/package/Microsoft.ML.OnnxRuntime/{ORT_VERSION}"


def step_export_onnx(tmp_dir: str) -> str:
    """Export the model to ONNX format using optimum."""
    print("\n[1/4] Exporting model to ONNX...")
    from optimum.onnxruntime import ORTModelForFeatureExtraction

    onnx_dir = os.path.join(tmp_dir, "onnx_model")
    model = ORTModelForFeatureExtraction.from_pretrained(
        MODEL_NAME, export=True
    )
    model.save_pretrained(onnx_dir)
    print(f"  -> Exported to {onnx_dir}")
    return onnx_dir


def step_quantize(onnx_dir: str, tmp_dir: str) -> str:
    """Quantize the ONNX model to INT8."""
    print("\n[2/4] Quantizing model to INT8...")
    from onnxruntime.quantization import quantize_dynamic, QuantType

    input_model = os.path.join(onnx_dir, "model.onnx")
    output_model = os.path.join(tmp_dir, "model.onnx")

    quantize_dynamic(
        input_model,
        output_model,
        weight_type=QuantType.QInt8,
    )

    orig_size = os.path.getsize(input_model) / (1024 * 1024)
    quant_size = os.path.getsize(output_model) / (1024 * 1024)
    print(f"  -> Original: {orig_size:.1f}MB -> Quantized: {quant_size:.1f}MB")
    return output_model


def step_get_tokenizer(onnx_dir: str, tmp_dir: str) -> str:
    """Copy tokenizer.json from the exported model."""
    print("\n[3/4] Copying tokenizer.json...")
    src = os.path.join(onnx_dir, "tokenizer.json")
    dst = os.path.join(tmp_dir, "tokenizer.json")
    if not os.path.exists(src):
        # Fallback: download from HuggingFace directly
        from transformers import AutoTokenizer
        tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
        tokenizer.save_pretrained(tmp_dir)
        src = os.path.join(tmp_dir, "tokenizer.json")
    else:
        shutil.copy2(src, dst)
    print(f"  -> tokenizer.json ready")
    return dst


def step_get_onnxruntime_dll(tmp_dir: str) -> str:
    """Download onnxruntime.dll from NuGet."""
    print("\n[4/4] Downloading onnxruntime.dll...")
    dll_path = os.path.join(tmp_dir, "onnxruntime.dll")

    # Download NuGet package (it's just a zip)
    nupkg_path = os.path.join(tmp_dir, "ort.nupkg")
    print(f"  -> Downloading from NuGet (v{ORT_VERSION})...")
    urllib.request.urlretrieve(ORT_NUGET_URL, nupkg_path)

    # Extract the DLL from the nupkg
    with zipfile.ZipFile(nupkg_path, "r") as z:
        dll_name = "runtimes/win-x64/native/onnxruntime.dll"
        if dll_name not in z.namelist():
            # List available files for debugging
            natives = [n for n in z.namelist() if "onnxruntime" in n.lower()]
            print(f"  -> Available natives: {natives}")
            raise FileNotFoundError(f"{dll_name} not found in nupkg")
        with z.open(dll_name) as src, open(dll_path, "wb") as dst:
            shutil.copyfileobj(src, dst)

    dll_size = os.path.getsize(dll_path) / (1024 * 1024)
    print(f"  -> onnxruntime.dll ({dll_size:.1f}MB)")
    os.remove(nupkg_path)
    return dll_path


def step_package_zip(model_path: str, tokenizer_path: str, dll_path: str):
    """Package everything into minilm.zip."""
    print(f"\n[Done] Packaging into {OUTPUT_ZIP}...")
    with zipfile.ZipFile(OUTPUT_ZIP, "w", zipfile.ZIP_DEFLATED) as zf:
        zf.write(model_path, "model.onnx")
        zf.write(tokenizer_path, "tokenizer.json")
        zf.write(dll_path, "onnxruntime.dll")

    total_size = os.path.getsize(OUTPUT_ZIP) / (1024 * 1024)
    print(f"  -> {OUTPUT_ZIP} ({total_size:.1f}MB)")
    print("\n[OK] Done! You can now run: go test -v -run E2E ./AI_translate")


def main():
    with tempfile.TemporaryDirectory(prefix="uml_model_") as tmp_dir:
        onnx_dir = step_export_onnx(tmp_dir)
        model_path = step_quantize(onnx_dir, tmp_dir)
        tokenizer_path = step_get_tokenizer(onnx_dir, tmp_dir)
        dll_path = step_get_onnxruntime_dll(tmp_dir)
        step_package_zip(model_path, tokenizer_path, dll_path)


if __name__ == "__main__":
    main()
