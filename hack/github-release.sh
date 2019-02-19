#!/bin/bash -e

CONFIG=$@

for line in $CONFIG; do
  eval "export ${line}"
done

# Define variables.
GH_API="https://api.github.com"
GH_REPO="$GH_API/repos/$owner/$repo"
GH_TAGS="$GH_REPO/releases/tags/$tag"
AUTH="Authorization: token $GITHUB_TOKEN"
WGET_ARGS="--content-disposition --auth-no-challenge --no-cookie"
CURL_ARGS="-LJO#"

if [[ "$tag" == 'LATEST' ]]; then
  GH_TAGS="$GH_REPO/releases/latest"
fi

# Validate token.
curl -o /dev/null -sH "$AUTH" $GH_REPO || { echo "Error: Invalid repo, token or network issue!";  exit 1; }

# check if we already have a release:
if curl --fail -o /dev/null -sH "$AUTH" $GH_TAGS; then
  # release exists - no need to create it. and no need to error
  exit 0
fi

BODY=$(cat <<EOF
{
  "tag_name": "${tag}",
  "target_commitish": "master",
  "name": "${tag}",
  "body": "${tag} release of squash binaries",
  "draft": false,
  "prerelease": false
}
EOF
)

curl --fail -d "${BODY}" -sH "$AUTH" -XPOST ${GH_REPO}/releases
