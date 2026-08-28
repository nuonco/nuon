import type { TCustomerManagedBundle } from '@/types'
import type { DiffSectionData } from '@/components/branches/AppConfigDiff'
import { BundleComparison } from './BundleComparison'

export default {
  title: 'Apps/Bundles/BundleComparison',
}

const digest = (character: string) => `sha256:${character.repeat(64)}`

const previousBundle = {
  id: 'bundle-previous',
  manifest_digest: digest('a'),
  artifacts: [
    {
      id: 'component-api-old',
      kind: 'component',
      logical_name: 'api',
      component_id: 'component-api',
      digest: digest('1'),
      config_digest: digest('2'),
      size: 1024,
    },
    {
      id: 'sandbox-old',
      kind: 'sandbox',
      logical_name: 'sandbox',
      digest: digest('3'),
      config_digest: digest('4'),
      size: 2048,
    },
  ],
} as TCustomerManagedBundle

const bundle = {
  id: 'bundle-current',
  manifest_digest: digest('b'),
  artifacts: [
    {
      id: 'component-api-new',
      kind: 'component',
      logical_name: 'api',
      component_id: 'component-api',
      digest: digest('5'),
      config_digest: digest('2'),
      size: 1536,
    },
    {
      id: 'sandbox-new',
      kind: 'sandbox',
      logical_name: 'sandbox',
      digest: digest('3'),
      config_digest: digest('4'),
      size: 2048,
    },
    {
      id: 'action-restart',
      kind: 'action_step',
      logical_name: 'restart/execute',
      digest: digest('6'),
      size: 512,
    },
  ],
} as TCustomerManagedBundle

const configSections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'api',
        op: 'change',
        componentType: 'job',
        fields: [
          {
            key: 'memory_size',
            op: 'change',
            diff: "'128' -> '256'",
          },
        ],
      },
    ],
  },
]

export const Default = () => (
  <BundleComparison
    bundle={bundle}
    previousBundle={previousBundle}
    configSections={configSections}
    orgId="org-demo"
    appId="app-demo"
  />
)

export const FirstBundle = () => (
  <BundleComparison bundle={bundle} orgId="org-demo" appId="app-demo" />
)
