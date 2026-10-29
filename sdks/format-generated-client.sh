# Sourced by sdk generate.sh scripts. SCRIPT_DIR and REPO_DIR must be set.
format_generated_client() {
  echo >&2 "formatting generated client"
  (
    cd "${REPO_DIR}/sdks/formatgen"
    go run . "${SCRIPT_DIR}/models" "${SCRIPT_DIR}/client"
  )
}
