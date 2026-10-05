#!/usr/bin/env bash
set -euo pipefail

INPUT_FILE="alert.json"
CHECK_ONLY=0

usage() {
  cat <<'USAGE'
Usage:
  check_vulnerabilities.sh [FILE]
  check_vulnerabilities.sh --ck [FILE]
  check_vulnerabilities.sh [FILE] --ck

Options:
  --ck    Check-only mode. Output only 0 or 1.
          0 = no vulnerability
          1 = vulnerability exists
  -h, --help
          Show this help.
USAGE
}

# Parse arguments.
# --ck can appear before or after the JSON file.
while [[ $# -gt 0 ]]; do
  case "$1" in
    --ck)
      CHECK_ONLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      echo "ERROR: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      INPUT_FILE="$1"
      shift
      ;;
  esac
done

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required." >&2
  exit 2
fi

if [[ ! -f "$INPUT_FILE" ]]; then
  echo "ERROR: file not found: $INPUT_FILE" >&2
  exit 2
fi

if ! jq empty "$INPUT_FILE" >/dev/null 2>&1; then
  echo "ERROR: invalid JSON: $INPUT_FILE" >&2
  exit 2
fi

if ! jq -e '.libraries | type == "array"' "$INPUT_FILE" >/dev/null; then
  echo "ERROR: expected .libraries to be an array" >&2
  exit 2
fi

# Extract vulnerabilities and de-duplicate by package + version + CVE.
UNIQUE_VULN_JSON=$(jq -c '
  [
    .libraries[] as $lib
    | ($lib.vulnerabilities // [])[]
    | {
        package: ($lib.groupId // $lib.artifactId // $lib.name // "UNKNOWN"),
        version: ($lib.version // "UNKNOWN"),
        cve: (.name // "UNKNOWN"),
        severity: ((.severity // .cvss3_severity // "UNKNOWN") | ascii_upcase),
        score: (.cvss3_score // .score // null)
      }
  ]
  | unique_by([.package, .version, .cve])
' "$INPUT_FILE")

UNIQUE_VULN_COUNT=$(jq 'length' <<< "$UNIQUE_VULN_JSON")

# ------------------------------------------------------------
# --ck mode
# Output ONLY:
#   0 = no vulnerability
#   1 = vulnerability exists
# Exit code is the same value.
# ------------------------------------------------------------
if (( CHECK_ONLY == 1 )); then
  if (( UNIQUE_VULN_COUNT > 0 )); then
    printf '1\n'
    exit 1
  else
    printf '0\n'
    exit 0
  fi
fi

# Normal report mode.
TOTAL_LIBS=$(jq '.libraries | length' "$INPUT_FILE")
VULN_LIB_ENTRIES=$(jq '[.libraries[] | select((.vulnerabilities // []) | length > 0)] | length' "$INPUT_FILE")
RAW_VULN_COUNT=$(jq '[.libraries[] | (.vulnerabilities // [])[]] | length' "$INPUT_FILE")
GLOBAL_UNIQUE_CVES=$(jq '[.[].cve] | unique | length' <<< "$UNIQUE_VULN_JSON")
VULNERABLE_PACKAGES=$(jq '[.[] | [.package, .version]] | unique | length' <<< "$UNIQUE_VULN_JSON")

get_count() {
  local severity="$1"
  jq --arg sev "$severity" '[.[] | select(.severity == $sev)] | length' <<< "$UNIQUE_VULN_JSON"
}

CRITICAL=$(get_count CRITICAL)
HIGH=$(get_count HIGH)
MEDIUM=$(get_count MEDIUM)
LOW=$(get_count LOW)
UNKNOWN=$(get_count UNKNOWN)

HIGHEST_SEVERITY=$(jq -r '
  def rank:
    if . == "CRITICAL" then 4
    elif . == "HIGH" then 3
    elif . == "MEDIUM" then 2
    elif . == "LOW" then 1
    else 0 end;

  if length == 0 then
    "NONE"
  else
    map(.severity) | unique | sort_by(rank) | reverse | .[0]
  end
' <<< "$UNIQUE_VULN_JSON")

if (( UNIQUE_VULN_COUNT > 0 )); then
  HAS_VULNERABILITY="YES"
else
  HAS_VULNERABILITY="NO"
fi

echo "========================================"
echo " Vulnerability Summary"
echo "========================================"
echo "File                    : $INPUT_FILE"
echo "Has vulnerabilities     : $HAS_VULNERABILITY"
echo "Highest severity        : $HIGHEST_SEVERITY"
echo "Library entries         : $TOTAL_LIBS"
echo "Vulnerable lib entries  : $VULN_LIB_ENTRIES"
echo "Vulnerable packages     : $VULNERABLE_PACKAGES"
echo "Raw vulnerability rows  : $RAW_VULN_COUNT"
echo "Unique global CVEs      : $GLOBAL_UNIQUE_CVES"
echo "Unique package/CVE count: $UNIQUE_VULN_COUNT"
echo
echo "Severity count (deduplicated by package+version+CVE)"
printf "  %-10s %d\n" "CRITICAL" "$CRITICAL"
printf "  %-10s %d\n" "HIGH" "$HIGH"
printf "  %-10s %d\n" "MEDIUM" "$MEDIUM"
printf "  %-10s %d\n" "LOW" "$LOW"
printf "  %-10s %d\n" "UNKNOWN" "$UNKNOWN"

echo
echo "========================================"
echo " Package Summary"
echo "========================================"
printf "%-28s %-12s %8s %10s\n" "PACKAGE" "VERSION" "VULNS" "MAX_LEVEL"
printf "%-28s %-12s %8s %10s\n" "-------" "-------" "-----" "---------"

jq -r '
  def rank:
    if . == "CRITICAL" then 4
    elif . == "HIGH" then 3
    elif . == "MEDIUM" then 2
    elif . == "LOW" then 1
    else 0 end;

  group_by([.package, .version])
  | map({
      package: .[0].package,
      version: .[0].version,
      count: length,
      maxSeverity: (map(.severity) | unique | sort_by(rank) | reverse | .[0])
    })
  | sort_by(.package, .version)
  | .[]
  | [.package, .version, (.count | tostring), .maxSeverity]
  | @tsv
' <<< "$UNIQUE_VULN_JSON" |
while IFS=$'\t' read -r package version count maxsev; do
  printf "%-28s %-12s %8s %10s\n" "$package" "$version" "$count" "$maxsev"
done

echo
echo "========================================"
echo " Vulnerability Detail"
echo "========================================"
printf "%-28s %-12s %-20s %-10s %s\n" "PACKAGE" "VERSION" "CVE" "SEVERITY" "SCORE"
printf "%-28s %-12s %-20s %-10s %s\n" "-------" "-------" "---" "--------" "-----"

jq -r '
  sort_by(.package, .version, .severity, .cve)
  | .[]
  | [.package, .version, .cve, .severity, ((.score // "-") | tostring)]
  | @tsv
' <<< "$UNIQUE_VULN_JSON" |
while IFS=$'\t' read -r package version cve severity score; do
  printf "%-28s %-12s %-20s %-10s %s\n" "$package" "$version" "$cve" "$severity" "$score"
done

# Normal mode also exposes vulnerability status through exit code.
if (( UNIQUE_VULN_COUNT > 0 )); then
  exit 1
fi
exit 0
