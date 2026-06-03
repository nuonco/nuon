import { useState } from 'react'
import { MatchPicker } from './MatchPicker'
import type { SubscriptionMatch } from './types'

export default { title: 'Match/MatchPicker' }

const Story = ({ initial }: { initial?: SubscriptionMatch }) => {
  const [value, setValue] = useState<SubscriptionMatch | undefined>(initial)
  return (
    <div className="max-w-lg p-4">
      <MatchPicker value={value} onChange={setValue} />
      <pre className="mt-4 text-xs bg-black/5 p-2 rounded">
        {JSON.stringify(value ?? null, null, 2)}
      </pre>
    </div>
  )
}

// Default: org-wide subscription. value is undefined; the "Everything in
// this org" radio is checked and no kind / predicate / entity controls
// are rendered.
export const Default = () => <Story />

// Preselected installs by ids — landing on the modal with a per-install
// scope already configured (mirrors the dashboard's edit flow).
export const PreselectedInstallIds = () => (
  <Story
    initial={{
      installs: { ids: ['inst_a', 'inst_b', 'inst_c'] },
    }}
  />
)

// Preselected components by labels — labels textinput is pre-filled with
// the canonical k=v,k=v string.
export const PreselectedComponentsByLabels = () => (
  <Story
    initial={{
      components: {
        selector: { match_labels: { env: 'prod', tier: 'critical' } },
      },
    }}
  />
)

// Preselected installs by exclusion only — "everything except env=stage".
// Include textinput stays empty, exclude textinput is pre-filled.
export const PreselectedInstallsExcludeOnly = () => (
  <Story
    initial={{
      installs: {
        selector: { not_match_labels: { env: 'stage' } },
      },
    }}
  />
)

// Preselected installs by both include + exclude — production tier=critical
// installs except the canary ones.
export const PreselectedInstallsIncludeAndExclude = () => (
  <Story
    initial={{
      installs: {
        selector: {
          match_labels: { env: 'prod' },
          not_match_labels: { canary: '*' },
        },
      },
    }}
  />
)

// Preselected actions with empty TargetMatch{} — predicate radio lands on
// "Any" so toggling kinds doesn't surprise the user.
export const PreselectedActionsAny = () => (
  <Story initial={{ actions: {} }} />
)

// Preselected component ids — exercises the app-first flow on edit. The
// app picker starts unselected; the existing chips render with bare ids
// until the user picks the owning app, at which point the listbox scopes
// to that app and the names resolve.
export const PreselectedComponentIds = () => (
  <Story initial={{ components: { ids: ['comp_a', 'comp_b'] } }} />
)

// Disabled — every control is greyed out and non-interactive.
export const Disabled = () => (
  <div className="max-w-lg p-4">
    <MatchPicker
      value={{ installs: { ids: ['inst_a'] } }}
      onChange={() => {}}
      disabled
    />
  </div>
)
