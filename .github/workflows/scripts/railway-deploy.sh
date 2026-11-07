#!/usr/bin/env bash
# Updates the Railway service source image to ${IMAGE_REPO}:${IMAGE_TAG_PREFIX}-${VERSION},
# then calls serviceInstanceDeploy to create a deployment from that source (not
# serviceInstanceRedeploy, which replays the previous deployment and can skip the new image).
# Auth: project-scoped token preferred; falls back to account/workspace token.
#
# Required env: VERSION, IMAGE_REPO, IMAGE_TAG_PREFIX,
#               RAILWAY_ENVIRONMENT_ID, RAILWAY_SERVICE_ID,
#               RAILWAY_PROJECT_TOKEN  (or  RAILWAY_TOKEN)

set -euo pipefail

if [ -z "${RAILWAY_ENVIRONMENT_ID:-}" ] || [ -z "${RAILWAY_SERVICE_ID:-}" ]; then
  echo "Missing RAILWAY_ENVIRONMENT_ID or RAILWAY_SERVICE_ID secrets." >&2
  exit 1
fi

if [ -n "${RAILWAY_PROJECT_TOKEN:-}" ]; then
  AUTH_ARGS=(-H "Project-Access-Token: ${RAILWAY_PROJECT_TOKEN}")
elif [ -n "${RAILWAY_TOKEN:-}" ]; then
  AUTH_ARGS=(-H "Authorization: Bearer ${RAILWAY_TOKEN}")
else
  echo "Set RAILWAY_TOKEN (account/workspace) or RAILWAY_PROJECT_TOKEN (project-scoped)." >&2
  exit 1
fi

IMAGE="${IMAGE_REPO}:${IMAGE_TAG_PREFIX}-${VERSION}"
ENDPOINT="https://backboard.railway.com/graphql/v2"

update_payload="$(jq -nc \
  --arg q 'mutation ($environmentId: String!, $serviceId: String!, $input: ServiceInstanceUpdateInput!) {
    serviceInstanceUpdate(environmentId: $environmentId, serviceId: $serviceId, input: $input)
  }' \
  --arg environmentId "${RAILWAY_ENVIRONMENT_ID}" \
  --arg serviceId "${RAILWAY_SERVICE_ID}" \
  --arg image "${IMAGE}" \
  '{query: $q, variables: {environmentId: $environmentId, serviceId: $serviceId, input: {source: {image: $image}}}}')"

echo "Updating Railway service source image to ${IMAGE}"
upd="$(curl -sS -X POST "${ENDPOINT}" \
  "${AUTH_ARGS[@]}" \
  -H "Content-Type: application/json" \
  -d "${update_payload}")"
echo "${upd}" | jq .
if echo "${upd}" | jq -e '.errors != null and (.errors | length) > 0' >/dev/null; then
  echo "serviceInstanceUpdate failed." >&2
  exit 1
fi

deploy_payload="$(jq -nc \
  --arg q 'mutation ($environmentId: String!, $serviceId: String!) {
    serviceInstanceDeploy(environmentId: $environmentId, serviceId: $serviceId)
  }' \
  --arg environmentId "${RAILWAY_ENVIRONMENT_ID}" \
  --arg serviceId "${RAILWAY_SERVICE_ID}" \
  '{query: $q, variables: {environmentId: $environmentId, serviceId: $serviceId}}')"

echo "Triggering Railway serviceInstanceDeploy (new deployment from updated source)"
dep="$(curl -sS -X POST "${ENDPOINT}" \
  "${AUTH_ARGS[@]}" \
  -H "Content-Type: application/json" \
  -d "${deploy_payload}")"
echo "${dep}" | jq .
if echo "${dep}" | jq -e '.errors != null and (.errors | length) > 0' >/dev/null; then
  echo "serviceInstanceDeploy failed." >&2
  exit 1
fi
