# NEXAGENT eBPF build.
#
# Generates Go bindings (bpf2go) and the compiled BPF object from the CO-RE
# source. Requires clang, llvm-strip, libbpf headers, and bpftool. See README.md.

CLANG ?= clang
BPF2GO ?= go run github.com/cilium/ebpf/cmd/bpf2go

.PHONY: generate vmlinux clean

generate: vmlinux
	$(BPF2GO) -cc $(CLANG) -target bpfel bpf bpf/tcp_rtt.bpf.c -- -I./bpf

vmlinux:
	@test -f bpf/vmlinux.h || bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/vmlinux.h

clean:
	rm -f bpf_bpfel.go bpf_bpfel.o bpf/vmlinux.h
