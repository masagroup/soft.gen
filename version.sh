#!/usr/bin/env bash
# Prints the version of the build, computed from the local Git state with the
# rules of the Go pseudo-versions, as sw.dos does. Only the tags of the major
# version of the module count (go.mod "module .../v2" -> tags v2.*), so that the
# v1 tags of the repository do not give a v1 version to the v2 module:
#   on a tag v2.1.0                -> v2.1.0
#   after a tag v2.1.0             -> v2.1.1-0.<yyyymmddhhmmss>-<hash>
#   after a pre-release v2.1.0-rc1 -> v2.1.0-rc1.0.<yyyymmddhhmmss>-<hash>
#   no v2 tag                      -> v2.0.0-<yyyymmddhhmmss>-<hash>
set -e

# 1. The major version of the module (v0 when the module path has no /vN suffix)
MAJOR=$(sed -n 's#^module .*/v\([0-9][0-9]*\)$#\1#p' go.mod 2>/dev/null || true)
MAJOR=${MAJOR:-0}

# 2. The UTC timestamp of the commit (yyyymmddhhmmss)
TIMESTAMP=$(TZ=UTC git log -1 --date='format-local:%Y%m%d%H%M%S' --format="%cd" 2>/dev/null || true)

# 3. The 12-character short commit hash
HASH=$(git rev-parse --short=12 HEAD 2>/dev/null || true)

# Fallback outside a git repository or without commits
if [ -z "$TIMESTAMP" ] || [ -z "$HASH" ]; then
    echo "v${MAJOR}.0.0-unknown"
    exit 0
fi

# 4. A tag of the major version on the commit itself is the version
EXACT_TAG=$(git describe --tags --exact-match --match "v${MAJOR}.*" 2>/dev/null || true)
if [ -n "$EXACT_TAG" ]; then
    echo "$EXACT_TAG"
    exit 0
fi

# 5. Otherwise the closest ancestor tag of the major version gives a pseudo-version
BASE_TAG=$(git describe --tags --abbrev=0 --match "v${MAJOR}.*" 2>/dev/null || true)
if [ -z "$BASE_TAG" ]; then
    # Rule 1: no base tag
    echo "v${MAJOR}.0.0-${TIMESTAMP}-${HASH}"
else
    case "$BASE_TAG" in
        *-*)
            # Rule 2: the base tag is a pre-release (v2.1.0-rc1)
            echo "${BASE_TAG}.0.${TIMESTAMP}-${HASH}"
            ;;
        *)
            # Rule 3: the base tag is a release (v2.1.0): the next patch version
            INC_TAG=$(echo "$BASE_TAG" | awk -F. '{OFS="."; $NF=$NF+1; print $0}')
            echo "${INC_TAG}-0.${TIMESTAMP}-${HASH}"
            ;;
    esac
fi
