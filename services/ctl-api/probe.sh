#!/bin/bash
export CGO_ENABLED=0
F=internal/pkg/workflows/workflow/activities/check_noop_plan.go
PKGS="$(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)"
echo "PROBE nproc=$(nproc) go=$(go version)"
cpu() { awk '/^cpu /{print $2+$3+$4+$7+$8, $2+$3+$4+$5+$6+$7+$8}' /proc/stat; }
probe() {
  name=$1; shift
  echo "var _probe$name = 1" >> $F
  read b0 t0 <<< "$(cpu)"; s=$(date +%s.%N)
  "$@" > /tmp/probe-$name.log 2>&1; rc=$?
  e=$(date +%s.%N); read b1 t1 <<< "$(cpu)"
  awk -v n="$name" -v rc=$rc -v s=$s -v e=$e -v b=$((b1-b0)) -v t=$((t1-t0)) -v np=$(nproc) -v ok=$(grep -c '^ok' /tmp/probe-$name.log) \
    'BEGIN{printf "PROBE %s rc=%d wall=%.0fs cores_busy=%.1f ok=%d\n", n, rc, e-s, np*b/t, ok}'
  grep -E '^(FAIL|---)' /tmp/probe-$name.log | head -5
}
probe A_testpkgs        env go test -vet=off $PKGS
probe B_p16             env go test -vet=off -p 16 $PKGS
probe C_ldw             env go test -vet=off -ldflags=-w $PKGS
probe D_p16_ldw         env go test -vet=off -p 16 -ldflags=-w $PKGS
probe E_p16_ldw_gogc400 env GOGC=400 go test -vet=off -p 16 -ldflags=-w $PKGS
probe F_old_all         env go test -vet=off ./...
