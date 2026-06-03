#!/usr/bin/env bash
# verify-ebpf-fleet.sh — gate NEXAGENT eBPF rollout on kernel/BTF/arch matrix.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EBPF_DIR="${ROOT}/backend/pkg/nexagent/ebpf"

echo "==> eBPF fleet matrix unit tests"
(cd "${EBPF_DIR}" && go test ./validate/... -count=1)

echo "==> host fleet probe (Linux BTF + kernel version)"
if ! (cd "${EBPF_DIR}" && go run ./cmd/fleet-validate); then
  echo "WARN: host probe failed — expected on non-Linux or kernels without BTF" >&2
  if [ "$(uname -s)" = "Linux" ]; then
    exit 1
  fi
fi

if [ "$(uname -s)" = "Linux" ] && [ -f /sys/kernel/btf/vmlinux ]; then
  echo "==> BTF vmlinux present ($(stat -c%s /sys/kernel/btf/vmlinux 2>/dev/null || stat -f%z /sys/kernel/btf/vmlinux) bytes)"
  if command -v bpftool >/dev/null 2>&1; then
    bpftool btf dump file /sys/kernel/btf/vmlinux format c 2>/dev/null | head -5 || true
  fi
  if command -v clang >/dev/null 2>&1 && command -v llvm-strip >/dev/null 2>&1; then
    echo "==> optional: bpf2go codegen (toolchain present)"
    (cd "${EBPF_DIR}" && make generate 2>/dev/null || echo "skip bpf2go (vmlinux.h may be missing in CI)")
    if [ -f "${EBPF_DIR}/bpf_bpfel.go" ]; then
      echo "==> compile ebpf loader (-tags ebpf)"
      (cd "${EBPF_DIR}" && go build -tags ebpf .)
    fi
  else
    echo "skip bpf compile (clang/llvm-strip not installed)"
  fi
else
  echo "skip BTF/bpf compile (not Linux or no BTF)"
fi

echo "eBPF fleet validation complete"
