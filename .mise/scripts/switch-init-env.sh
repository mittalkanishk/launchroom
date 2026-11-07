#!/usr/bin/env bash
set -euo pipefail

if [ -f .env ]; then echo ".env already exists; delete it to regenerate" >&2; exit 1; fi

secret() { openssl rand -hex 24; }

cat > .env <<EOF
DB_NAME=switch
DB_OWNER_USER=postgres
DB_OWNER_PASSWORD=$(secret)
DB_USER=switch_app
DB_PASSWORD=$(secret)
AGENT_REGISTRATION_TOKEN=$(secret)
SECRET_KEYS=local:$(openssl rand -hex 32)
GATEWAY_ADMIN_EMAIL=admin@switch.local
GATEWAY_ADMIN_PASSWORD=$(secret)
FRONTEND_BASE_URL=http://localhost:13000
CLOUDFLARE_TUNNEL_TOKEN=
MESSAGING_PUBLIC_URL=
CONTROLLER_TOKEN_SECRET=$(openssl rand -hex 32)
SWITCH_CONTROLLER_ENROLL_CODE=
CLAUDE_CODE_OAUTH_TOKEN=
OPENAI_API_KEY=
SITES_DOMAIN=
FORMSPREE_PROJECT_ID=
FORMSPREE_DEPLOY_KEY=
EOF

echo "wrote .env; gateway login is admin@switch.local with GATEWAY_ADMIN_PASSWORD from .env"
echo "fill in CLOUDFLARE_TUNNEL_TOKEN and MESSAGING_PUBLIC_URL from your Cloudflare tunnel,"
echo "CLAUDE_CODE_OAUTH_TOKEN from \`claude setup-token\`, and SWITCH_CONTROLLER_ENROLL_CODE from Machines > Add machine"
