// SPDX-License-Identifier: GPL-2.0
//
// tcp_rtt.bpf.c — NEXAGENT eBPF CO-RE program.
//
// Attaches to the tcp_rcv_established tracepoint/kprobe and records smoothed
// round-trip time (srtt) per established TCP connection into a histogram map.
// Userspace (loader.go) reads the histogram and converts buckets into the same
// collector.Sample shape used by the procfs and host collectors, so the eBPF
// path is a drop-in replacement for the procfs KernelCollector on kernels with
// BTF + CO-RE support (Linux 5.4+ with CONFIG_DEBUG_INFO_BTF=y).
//
// Build (requires clang >= 12, libbpf, bpftool):
//   make -C pkg/nexagent/ebpf
// which runs bpf2go to generate the Go bindings consumed by loader.go.

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

#define MAX_SLOTS 27

struct hist_key {
    __u32 saddr;
    __u32 daddr;
};

struct hist {
    __u64 slots[MAX_SLOTS];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, struct hist_key);
    __type(value, struct hist);
} rtt_hist SEC(".maps");

// log2l approximates a power-of-two bucket index for histogram slotting.
static __always_inline __u64 log2l(__u64 v) {
    __u64 r = 0;
    while (v >>= 1) {
        r++;
    }
    return r;
}

SEC("kprobe/tcp_rcv_established")
int BPF_KPROBE(handle_tcp_rcv, struct sock *sk) {
    if (!sk) {
        return 0;
    }

    struct tcp_sock *ts = (struct tcp_sock *)sk;
    __u32 srtt = BPF_CORE_READ(ts, srtt_us) >> 3; // srtt_us is in 1/8 us units

    struct hist_key key = {};
    key.saddr = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
    key.daddr = BPF_CORE_READ(sk, __sk_common.skc_daddr);

    struct hist *h = bpf_map_lookup_elem(&rtt_hist, &key);
    if (!h) {
        struct hist zero = {};
        bpf_map_update_elem(&rtt_hist, &key, &zero, BPF_NOEXIST);
        h = bpf_map_lookup_elem(&rtt_hist, &key);
        if (!h) {
            return 0;
        }
    }

    __u64 slot = log2l(srtt);
    if (slot >= MAX_SLOTS) {
        slot = MAX_SLOTS - 1;
    }
    __sync_fetch_and_add(&h->slots[slot], 1);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";
