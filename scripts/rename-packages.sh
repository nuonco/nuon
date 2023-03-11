#!/usr/bin/env bash

set -e
set -o pipefail
set -u

# update references to protos
#go-imports-rename --save 'github.com/powertoolsdev/protos/workflows => github.com/powertoolsdev/mono/pkg/protos/workflows'
#go-imports-rename --save 'github.com/powertoolsdev/protos/orgs-api => github.com/powertoolsdev/mono/pkg/protos/orgs-api'
#go-imports-rename --save 'github.com/powertoolsdev/protos/api => github.com/powertoolsdev/mono/pkg/protos/api'
#go-imports-rename --save 'github.com/powertoolsdev/protos/components => github.com/powertoolsdev/mono/pkg/protos/components'
#go-imports-rename --save 'github.com/powertoolsdev/protos/shared => github.com/powertoolsdev/mono/pkg/protos/shared'

## update individual services internal directories
#go-imports-rename --save 'github.com/powertoolsdev/api/internal => github.com/powertoolsdev/mono/services/api/internal'
#go-imports-rename --save 'github.com/powertoolsdev/orgs-api/internal => github.com/powertoolsdev/mono/services/orgs-api/internal'
#go-imports-rename --save 'github.com/powertoolsdev/nuonctl/internal => github.com/powertoolsdev/mono/services/nuonctl/internal'
#go-imports-rename --save 'github.com/powertoolsdev/workers-apps/internal => github.com/powertoolsdev/mono/services/workers-apps/internal'
#go-imports-rename --save 'github.com/powertoolsdev/workers-deployments/internal => github.com/powertoolsdev/mono/services/workers-deployments/internal'
#go-imports-rename --save 'github.com/powertoolsdev/workers-instances/internal => github.com/powertoolsdev/mono/services/workers-instances/internal'
#go-imports-rename --save 'github.com/powertoolsdev/workers-installs/internal => github.com/powertoolsdev/mono/services/workers-installs/internal'
#go-imports-rename --save 'github.com/powertoolsdev/workers-orgs/internal => github.com/powertoolsdev/mono/services/workers-orgs/internal'
#go-imports-rename --save 'github.com/powertoolsdev/workers-executors/internal => github.com/powertoolsdev/mono/services/workers-executors/internal'

## update service cmd directories
#go-imports-rename --save 'github.com/powertoolsdev/api/cmd => github.com/powertoolsdev/mono/services/api/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/orgs-api/cmd => github.com/powertoolsdev/mono/services/orgs-api/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/nuonctl/cmd => github.com/powertoolsdev/mono/services/nuonctl/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/workers-apps/cmd => github.com/powertoolsdev/mono/services/workers-apps/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/workers-deployments/cmd => github.com/powertoolsdev/mono/services/workers-deployments/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/workers-instances/cmd => github.com/powertoolsdev/mono/services/workers-instances/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/workers-installs/cmd => github.com/powertoolsdev/mono/services/workers-installs/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/workers-orgs/cmd => github.com/powertoolsdev/mono/services/workers-orgs/cmd'
#go-imports-rename --save 'github.com/powertoolsdev/workers-executors/cmd => github.com/powertoolsdev/mono/services/workers-executors/cmd'

## update packages
#go-imports-rename --save 'github.com/powertoolsdev/go-aws-assume-role => github.com/powertoolsdev/mono/pkg/aws-assume-role'
#go-imports-rename --save 'github.com/powertoolsdev/go-common => github.com/powertoolsdev/mono/pkg/common'
#go-imports-rename --save 'github.com/powertoolsdev/go-config => github.com/powertoolsdev/mono/pkg/config'
#go-imports-rename --save 'github.com/powertoolsdev/go-components => github.com/powertoolsdev/mono/pkg/components'
#go-imports-rename --save 'github.com/powertoolsdev/go-fetch => github.com/powertoolsdev/mono/pkg/fetch'
#go-imports-rename --save 'github.com/powertoolsdev/go-generics => github.com/powertoolsdev/mono/pkg/generics'
#go-imports-rename --save 'github.com/powertoolsdev/go-helm => github.com/powertoolsdev/mono/pkg/helm'
#go-imports-rename --save 'github.com/powertoolsdev/go-kube => github.com/powertoolsdev/mono/pkg/kube'
#go-imports-rename --save 'github.com/powertoolsdev/go-sender => github.com/powertoolsdev/mono/pkg/sender'
#go-imports-rename --save 'github.com/powertoolsdev/go-shared-types => github.com/powertoolsdev/mono/pkg/shared-types'
#go-imports-rename --save 'github.com/powertoolsdev/go-terraform => github.com/powertoolsdev/mono/pkg/terraform'
#go-imports-rename --save 'github.com/powertoolsdev/go-workflows-meta => github.com/powertoolsdev/mono/pkg/workflows-meta'

# TODO(jm): update this once all users of go-waypoint are using v2
#go-imports-rename --save 'github.com/powertoolsdev/go-waypoint => github.com/powertoolsdev/mono/pkg/waypoint'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/components/generated/types => github.com/powertoolsdev/mono/pkg/types/components'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/executors/v1 => github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/deployments/v1 => github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/apps/v1 => github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/orgs/v1 => github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/instances/v1 => github.com/powertoolsdev/mono/pkg/types/workflows/instances/v1'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/installs/v1 => github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/types/shared/status/v1 => github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1'

go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/shared/generated => github.com/powertoolsdev/mono/pkg/types/shared'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/api/generated => github.com/powertoolsdev/mono/pkg/types/api'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/orgs-api/generated => github.com/powertoolsdev/mono/pkg/types/orgs-api'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/protos/components/generated => github.com/powertoolsdev/mono/pkg/types/components'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/types/api/types => github.com/powertoolsdev/mono/pkg/types/api'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/types/shared/types => github.com/powertoolsdev/pkg/types/shared'
go-imports-rename --save 'github.com/powertoolsdev/pkg => github.com/powertoolsdev/mono/pkg'
go-imports-rename --save 'github.com/powertoolsdev/mono/pkg/types/orgs-api/types => github.com/powertoolsdev/mono/pkg/types/orgs-api'
