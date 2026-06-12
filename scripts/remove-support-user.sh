#!/usr/bin/env bash
set -euo pipefail

# Remove a (former) support user from every org they were added to.
# Mirrors authz.RemoveAccountOrgRoles: hard-deletes account_roles + org_invites.
#
# Usage:
#   ./scripts/remove-support-user.sh stephen@nuon.co            # dry run (stage)
#   ./scripts/remove-support-user.sh stephen@nuon.co --apply    # delete (stage)
#   PROFILE=prod.NuonAdmin ./scripts/remove-support-user.sh stephen@nuon.co --apply
#
# Requires: Twingate VPN, aws cli, jq, psql

EMAIL="${1:?usage: $0 <email> [--apply]}"
APPLY="${2:-}"

PROFILE="${PROFILE:-stage.NuonPowerUser}"
DB_ID="${DB_ID:-primary-ctl-api}"

DB_HOST=$(aws --profile "$PROFILE" rds describe-db-instances --db-instance-identifier="$DB_ID" \
  | jq -r '.DBInstances[0].Endpoint.Address')

PGPASSWORD=$(aws --profile "$PROFILE" rds generate-db-auth-token \
  --hostname "$DB_HOST" --port 5432 --username ctl_api)
export PGPASSWORD

PSQL=(psql -h "$DB_HOST" -U ctl_api ctl_api -v email="$EMAIL")

echo "== account =="
"${PSQL[@]}" -c "select id, email, account_type, created_at from accounts where email = :'email';"

echo "== org roles held =="
"${PSQL[@]}" -c "
select o.id as org_id, o.name as org_name, r.role_type, ar.created_at
from account_roles ar
join accounts a on a.id = ar.account_id
join roles r on r.id = ar.role_id
left join orgs o on o.id = ar.org_id
where a.email = :'email'
order by o.name;"

echo "== pending org invites =="
"${PSQL[@]}" -c "select org_id, email, created_at from org_invites where email = :'email';"

if [[ "$APPLY" != "--apply" ]]; then
  echo
  echo "dry run — re-run with --apply to delete the rows above"
  exit 0
fi

"${PSQL[@]}" <<'SQL'
begin;
delete from account_roles where account_id in (select id from accounts where email = :'email');
delete from org_invites where email = :'email';
commit;
SQL

echo "done — removed all org roles and invites for $EMAIL"
