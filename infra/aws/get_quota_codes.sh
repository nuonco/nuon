#!/usr/bin/env bash

# This script will query AWS for the list of quota codes for the
# service code passed as the only parameter to this script.

# To update / add codes to the enum, run this script and paste the
# output into the enum. Then run `npm run lint:fix` repeatedly until
# it is in alphabetical order.

while read -r code name;
do
    nrml=$(echo "$name" | sed 's/[- (),]/_/g' | sed 's/__/_/g')
    printf "%s = \"%s\",\n" "${nrml^^}" "$code";
done < <(
    aws service-quotas list-service-quotas --service-code "$1" \
    | jq -rc '.Quotas[] | [ .QuotaCode, .QuotaName ] | @tsv'
)
