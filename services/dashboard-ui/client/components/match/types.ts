// TypeScript mirror of the Go contract in pkg/labels/match.go.
//
// Wire format uses snake_case, exactly as the Go struct tags emit it. The
// backend stamps `match` as `swaggertype:"object"`, so the auto-generated
// SDK shape is a generic object — these hand-written types are the
// canonical shape everything client-side reads / writes.
//
// Stay in lockstep with pkg/labels/match.go (TargetKind, TargetMatch,
// SubscriptionMatch). Composition rules mirror the Go matcher exactly:
//   - SubscriptionMatch.Matches  : OR across non-nil kinds
//   - TargetMatch.matches        : OR within (id ∈ IDs OR Selector hit)
//   - Selector.Matches           : AND across MatchLabels
//
// A nil/undefined SubscriptionMatch is the org-wide subscription — match
// every event in the linked org.

// Labels mirrors pkg/labels.Labels — a flat key/value map. Values may be
// the wildcard `"*"` (matches any value for the key).
export type Labels = Record<string, string>

// TargetKind names the resource taxonomy a TargetMatch applies to. Kept
// deliberately narrow — only kinds the picker UI exposes today. NOT the
// same set as interests.ResourceKind: pkg/labels lives below internal/ in
// the import graph and the resource taxonomy is a property of the
// dispatch layer, not of label matching.
export type TargetKind = 'installs' | 'components' | 'actions' | 'app_branches'

// Selector mirrors pkg/labels.Selector. Both MatchLabels and
// NotMatchLabels values may be a literal (`env=prod`) or the wildcard
// `"*"` (matches any value for the key — the bare-key form `env`
// round-trips as `env=*`). NotMatchLabels lets users say "everything
// except env=stage" without enumerating every other env value;
// composition rule is AND across both maps (entity matches iff every
// key in match_labels matches AND no key in not_match_labels matches).
export interface Selector {
  match_labels?: Labels
  not_match_labels?: Labels
}

// TargetMatch filters one entity kind. An empty TargetMatch ({}) is
// intentionally valid and means "any entity of this kind" — the modal
// needs an explicit "Any installs" option distinct from "Specific
// installs" / "By labels" without forcing a sentinel value.
export interface TargetMatch {
  ids?: string[]
  selector?: Selector
}

// SubscriptionMatch is the per-subscription routing filter persisted as
// JSONB on slack_channel_subscriptions.match. Schema supports populating
// multiple kinds in one row (CLI uses this); the v1 dashboard / Slack
// modal only ever writes one kind at a time.
export interface SubscriptionMatch {
  installs?: TargetMatch
  components?: TargetMatch
  actions?: TargetMatch
  app_branches?: TargetMatch
}

// Canonical, ordered list of target kinds. Mirrors the Slack subscribe
// modal's `Resource type` radio order.
export const ALL_TARGET_KINDS: TargetKind[] = [
  'installs',
  'components',
  'actions',
  'app_branches',
]

// Sentence-case display labels per kind. Mirror Slack modal's
// targetKindLabel(_, plural=true) for visual parity.
export const TARGET_KIND_LABELS: Record<TargetKind, string> = {
  installs: 'Installs',
  components: 'Components',
  actions: 'Actions',
  app_branches: 'App branches',
}

// Singular forms for prose like "Any install" / "Match by".
export const TARGET_KIND_LABELS_SINGULAR: Record<TargetKind, string> = {
  installs: 'install',
  components: 'component',
  actions: 'action',
  app_branches: 'app branch',
}

// Plural forms (lowercase) for prose like "Search installs", "Match every component in this org".
export const TARGET_KIND_LABELS_PLURAL: Record<TargetKind, string> = {
  installs: 'installs',
  components: 'components',
  actions: 'actions',
  app_branches: 'app branches',
}

// Predicate is the per-kind matcher selector exposed in the picker UI.
// Mirrors the predicate radio in the Slack subscribe modal.
export type Predicate = 'any' | 'specific' | 'labels'

// firstPopulatedKind walks ALL_TARGET_KINDS in order and returns the
// first one whose TargetMatch slot is set on the SubscriptionMatch. Used
// by describeMatch and to seed the kind radio when editing existing rows
// — mirrors the Slack modal's behaviour, which also surfaces only the
// first populated kind.
export const firstPopulatedKind = (
  m: SubscriptionMatch | undefined
): TargetKind | undefined => {
  if (!m) return undefined
  for (const k of ALL_TARGET_KINDS) {
    if (m[k]) return k
  }
  return undefined
}

// describeMatch returns a short human-readable summary of the match
// predicate. Mirrors the Go describeMatch helper in
// services/ctl-api/internal/app/slack/service/subscribe_modal.go so the
// dashboard table and the Slack modal use the same vocabulary.
//
//   - undefined  / no kind populated → "Org-wide"
//   - include only                   → "Components: env=prod, owner=*"
//   - exclude only                   → "Components: not env=stage"
//   - both                           → "Components: env=prod; not env=stage"
//   - ids                            → "3 installs"
//   - empty TargetMatch{}            → "Any installs"
export const describeMatch = (m: SubscriptionMatch | undefined): string => {
  const kind = firstPopulatedKind(m)
  if (!m || !kind) return 'Org-wide'
  const tm = m[kind]
  if (!tm) return 'Org-wide'
  const sel = tm.selector
  const inc = sel?.match_labels
  const exc = sel?.not_match_labels
  const incNonEmpty = inc && Object.keys(inc).length > 0
  const excNonEmpty = exc && Object.keys(exc).length > 0
  if (incNonEmpty || excNonEmpty) {
    const parts: string[] = []
    if (incNonEmpty) parts.push(labelsToQueryString(inc))
    if (excNonEmpty) parts.push(`not ${labelsToQueryString(exc)}`)
    return `${TARGET_KIND_LABELS[kind]}: ${parts.join('; ')}`
  }
  if (tm.ids && tm.ids.length > 0) {
    const noun =
      tm.ids.length === 1
        ? TARGET_KIND_LABELS_SINGULAR[kind]
        : TARGET_KIND_LABELS_PLURAL[kind]
    return `${tm.ids.length} ${noun}`
  }
  return `Any ${TARGET_KIND_LABELS_PLURAL[kind]}`
}

// labelsToQueryString joins a Labels map back into the
// k=v,k=*-grammar text the labels textinput consumes. Keys are sorted
// so the rendered string is deterministic across re-renders. Wildcard
// values ("*") render as `k=*` to match the Slack modal's
// labelsToQueryString helper.
export const labelsToQueryString = (m: Labels | undefined): string => {
  if (!m) return ''
  const keys = Object.keys(m).sort()
  return keys.map((k) => `${k}=${m[k]}`).join(', ')
}
