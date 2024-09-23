// clickhouse s3 backup crons
// docs: https://clickhouse.com/docs/en/operations/backup#configuring-backuprestore-to-use-an-s3-endpoint
//
// NOTE: we use the S3() method so we can specify the path even though the disk itself specifies a path /tables

resource "kubectl_manifest" "clickhouse_backup_script" {
  yaml_body = yamlencode({
    "apiVersion" = "v1"
    "kind" = "ConfigMap"
    "metadata" = {
      "name" = "clickhouse-backup-to-s3-script"
      "managed" = "terraform"
    }
    "data" = {
      "backup.sh" = file("${path.module}/backup.sh")
    }
  })
  depends_on = [
    kubectl_manifest.clickhouse_installation
  ]
}

// we make a cron for each of the tables in locals.tables
// TODO(fd): running n crons where n = len(locals.tables) is likely to tax the db as we grow.
// consider making this a job and triggering it via ctl-api or some other process w/ insight
// into the state of these tables so the system can choose when to back itself up. this would mean we
// woudl backup more often during on-hours and less during off-hours, presumably saving resources.
resource "kubectl_manifest" "clickhouse_backup_crons" {
  for_each = toset(local.tables)

  yaml_body = yamlencode({
    "apiVersion" = "batch/v1"
    "kind"       = "CronJob"
    "metadata"   = {
      "name"      = "clickhouse-backup-to-s3-${each.key}"
      "namespace" = "clickhouse"
      "table"     = each.key
    }
    "spec" = {
      "jobTemplate" = {
        "spec" = {
          "template" = {
            "spec" = {
              "containers" = [
                {
                  "command" = [
                    "bash",
                    "/usr/bin/backup.sh",
                    "${each.key}",
                  ]
                  "env" = [
                    {
                      "name"  = "BUCKET_URL"
                      "value" = "https://${module.bucket.s3_bucket_bucket_domain_name}"
                    },
                  ]
                  "image"           = "clickhouse/clickhouse-server:${local.image_tag}"
                  "imagePullPolicy" = "IfNotPresent"
                  "name"            = "clickhouse-backup-to-s3"
                  "volumeMounts"    = [
                    {
                      "mountPath" = "/usr/bin"
                      "name"      = "config-volume"
                    },
                  ]
                },
              ]
              "restartPolicy"      = "OnFailure"
              "serviceAccountName" = "default"
              "volumes"            = [
                {
                  "configMap" = {
                    "name" = "clickhouse-backup-to-s3-script"
                  }
                  "name" = "config-volume"
                },
              ]
            }
          }
        }
      }
      "schedule" = "13,37,51 */1 * * *" // TODO(fd): add some randomness +/- 3 min
    }
  })
}
