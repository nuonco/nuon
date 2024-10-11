resource "kubectl_manifest" "storageclass_ebi" {
  yaml_body = yamlencode({
    "allowedTopologies" = [
      {
        "matchLabelExpressions" = [
          {
            "key"    = "topology.ebs.csi.aws.com/zone"
            "values" = local.availability_zones
          },
        ]
      },
    ]
    "apiVersion" = "storage.k8s.io/v1"
    "kind"       = "StorageClass"
    "metadata" = {
      "name" = "ebi"
    }
    "parameters" = {
      "fsType" = "ext4"
      "type"   = "gp2"
    }
    "provisioner"       = "kubernetes.io/aws-ebs"
    "reclaimPolicy"     = "Delete"
    "volumeBindingMode" = "WaitForFirstConsumer"
  })
}
