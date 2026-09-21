export default {
  title: 'Diffs/Diff',
}

import { Diff } from './Diff'
import {
  deploymentAfter,
  deploymentBefore,
  longManifest,
  scatteredManifest,
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

export const Chunking = () => (
  <div className="p-4">
    <Diff
      before={scatteredManifest(false)}
      after={scatteredManifest(true)}
      language="yaml"
      filename="settings.yaml"
      maxHeight={900}
    />
  </div>
)

export const ChunkingSplit = () => (
  <div className="p-4">
    <Diff
      before={scatteredManifest(false)}
      after={scatteredManifest(true)}
      language="yaml"
      filename="settings.yaml"
      view="split"
      maxHeight={900}
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
