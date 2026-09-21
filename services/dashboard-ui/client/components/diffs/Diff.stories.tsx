export default {
  title: 'Diffs/Diff',
}

import { Diff } from './Diff'
import {
  deploymentAfter,
  deploymentBefore,
  longManifest,
  terraformAfter,
  terraformBefore,
} from './__fixtures__/diffs'

export const Unified = () => (
  <div className="p-4">
    <Diff
      before={deploymentBefore}
      after={deploymentAfter}
      language="yaml"
      filename="deployment.yaml"
    />
  </div>
)

export const Split = () => (
  <div className="p-4">
    <Diff
      before={deploymentBefore}
      after={deploymentAfter}
      language="yaml"
      filename="deployment.yaml"
      view="split"
    />
  </div>
)

export const Terraform = () => (
  <div className="p-4">
    <Diff
      before={terraformBefore}
      after={terraformAfter}
      language="terraform"
      filename="iam.tf"
    />
  </div>
)

export const LongWithCollapsedContext = () => (
  <div className="p-4">
    <Diff
      before={longManifest(2)}
      after={longManifest(6)}
      language="yaml"
      filename="settings.yaml"
    />
  </div>
)

export const NoChrome = () => (
  <div className="p-4">
    <Diff
      before={deploymentBefore}
      after={deploymentAfter}
      language="yaml"
      search={false}
    />
  </div>
)
