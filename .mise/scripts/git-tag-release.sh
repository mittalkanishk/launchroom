#!/usr/bin/env bash
set -euo pipefail

# Use version-sorted tags, not `git describe`: describe picks the nearest tag on HEAD's ancestry
# and is arbitrary when several tags share one commit, so it can show an older tag than the latest.
latest_tag=$(git tag -l 'v*' --sort=-v:refname | head -n 1)
echo "Latest tag is ${latest_tag:-none}"
read -r -p "Enter new release tag (with a v e.g. v1.2.3): " new_tag
git tag "$new_tag"
git push origin "$new_tag"
echo "Released new tag $new_tag"
