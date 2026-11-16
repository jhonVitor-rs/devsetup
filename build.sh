# build.sh
set -euo pipefail

VERSION="${VERSION:-dev}"
OUT=dist
rm -rf "$OUT" && mkdir -p "$OUT"

targets=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

for t in "${targets[@]}"; do
  os="${t%/*}"; arch="${t#*/}"
  ext=""; [ "$os" = "windows" ] && ext=".exe"
  name="devsetup-${os}-${arch}${ext}"
  echo "→ $name"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath \
      -ldflags "-s -w -X github.com/jhonVitor-rs/devsetup/cmd.version=${VERSION}" \
      -o "$OUT/$name" .
done

(cd "$OUT" && sha256sum * > checksums.txt)
echo "pronto: $OUT/"
