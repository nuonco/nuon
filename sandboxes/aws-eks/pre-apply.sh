#!/bin/sh

set -u
set -o pipefail

echo "executing error-destroy script"

echo "ensuring AWS is setup"
aws sts get-caller-identity > /dev/null

echo "looking for ENIs which were orphaned by vpc-cni plugin"
ENIS=$(aws ec2 \
  describe-network-interfaces \
  --filters Name=tag:cluster.k8s.amazonaws.com/name,Values=$NUON_INSTALL_ID)

echo $ENIS | jq -r '.NetworkInterfaces[].NetworkInterfaceId' | while read -r eni_id ; do
  echo "deleting ENI $eni_id"
  aws ec2 delete-network-interface --network-interface-id=$eni_id
done
