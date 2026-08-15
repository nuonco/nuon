#!/bin/sh
# actions-supervisor: mounted into an image-backed action container and run via
# the image's own /bin/sh. Installs the nuon_output helper on PATH, then runs
# the rendered step script in workdir and propagates its exit code.
set -u

script=""
workdir="."
while [ "$#" -gt 0 ]; do
	case "$1" in
	--script)
		script="${2:-}"
		shift 2
		;;
	--workdir)
		workdir="${2:-}"
		shift 2
		;;
	*)
		shift
		;;
	esac
done

if [ -z "$script" ]; then
	echo "actions-supervisor: --script is required" >&2
	exit 1
fi

root="${NUON_ACTIONS_ROOT:-$workdir}"

# The helper is supervisor-private, so it goes in the container's own filesystem
# rather than the shared workspace: a non-root image often cannot write the bind
# mount, and silently skipping the helper would leave nuon_output missing at the
# point a step calls it.
bindir="${TMPDIR:-/tmp}/nuon-actions-bin"
if ! mkdir -p "$bindir" 2>/dev/null; then
	echo "actions-supervisor: unable to create $bindir for the nuon_output helper" >&2
	exit 1
fi

# nuon_output <key> <value>: append a key=value line to the outputs file, which
# the runner reads back from the shared workspace mount after the container exits.
if ! cat >"$bindir/nuon_output" <<'HELPER'
#!/bin/sh
printf '%s=%s\n' "$1" "$2" >>"$NUON_ACTIONS_OUTPUT_FILEPATH"
HELPER
then
	echo "actions-supervisor: unable to install the nuon_output helper" >&2
	exit 1
fi
chmod 0755 "$bindir/nuon_output"
PATH="$bindir:$PATH"
export PATH

# ensure the step script carries a shebang so it runs under the image's shell
case "$(head -n1 "$script" 2>/dev/null)" in
'#!'*) ;;
*)
	tmp="$script.nuon-tmp"
	{
		printf '#!/bin/sh\n'
		cat "$script"
	} >"$tmp" && mv "$tmp" "$script"
	;;
esac
chmod 0755 "$script" 2>/dev/null || true

cd "$workdir" || {
	echo "actions-supervisor: unable to cd to $workdir" >&2
	exit 1
}

"$script"
exit "$?"
