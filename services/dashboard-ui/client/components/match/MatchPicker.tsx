import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Select } from '@/components/common/form/Select'
import { useOrg } from '@/hooks/use-org'
import { getApps } from '@/lib'
import { cn } from '@/utils/classnames'
import { EntityMultiSelect } from './EntityMultiSelect'
import { labelsToQueryString, parseLabelsQuery } from './parse'
import {
  ALL_TARGET_KINDS,
  TARGET_KIND_LABELS,
  TARGET_KIND_LABELS_PLURAL,
  TARGET_KIND_LABELS_SINGULAR,
  firstPopulatedKind,
  type Predicate,
  type SubscriptionMatch,
  type TargetKind,
} from './types'

type MatchMode = 'all' | 'specific'

// MatchPicker mirrors the Slack /nuon subscribe modal's scope picker
// (buildSubscribeModalView in services/ctl-api/internal/app/slack/service/
// subscribe_modal.go) so the dashboard and Slack surfaces feel
// identical:
//
//   ◉ Everything in this org   ○ Specific resources
//   (when Specific)
//     Resource type:  ◉ Installs ○ Components ○ Actions
//     (Components/Actions only) App: [ searchable app picker ]
//     Match by:       ◉ Any ○ Specific ○ Labels
//     (Specific)      [ entity multi-select scoped to the picked app
//                       — installs stay org-scoped ]
//     (Labels)        [ "env=prod, tier=critical, owner=*" include input ]
//                     [ "env=stage" exclude input — "everything except…" ]
//
// `value` is the exact wire shape persisted to the slack_channel_subscriptions
// row's `match` JSONB column. Undefined => org-wide subscription.
//
// Components and actions are app-owned, so picking them as a kind
// gates the entity multi-select on first picking an app — we use the
// per-app endpoints (GET /v1/apps/:app_id/{components,actions}) which
// already support q + labels. Installs stay org-scoped (they have an
// app_id but enumerating across the org is the right UX for them).
//
// The picker keeps internal state for the currently-edited kind /
// predicate / labels text / app id so switching kinds or predicates
// doesn't lose in-flight selections from the other branches. Only the
// externally visible `value` ever leaks back through onChange.
export const MatchPicker = ({
  value,
  onChange,
  disabled,
}: {
  value: SubscriptionMatch | undefined
  onChange: (next: SubscriptionMatch | undefined) => void
  disabled?: boolean
}) => {
  const initialKind = firstPopulatedKind(value) ?? 'installs'
  const initialMode: MatchMode = value ? 'specific' : 'all'
  const initialPredicate: Predicate = derivePredicate(value, initialKind)
  const initialLabels = useMemo(() => {
    const tm = value?.[initialKind]
    return labelsToQueryString(tm?.selector?.match_labels)
  }, [value, initialKind])
  const initialExcludeLabels = useMemo(() => {
    const tm = value?.[initialKind]
    return labelsToQueryString(tm?.selector?.not_match_labels)
  }, [value, initialKind])

  const [mode, setMode] = useState<MatchMode>(initialMode)
  const [kind, setKind] = useState<TargetKind>(initialKind)
  const [predicate, setPredicate] = useState<Predicate>(initialPredicate)
  const [labelsRaw, setLabelsRaw] = useState<string>(initialLabels)
  const [excludeLabelsRaw, setExcludeLabelsRaw] =
    useState<string>(initialExcludeLabels)
  // appId is only meaningful when kind ∈ {components, actions} and
  // predicate is 'specific'. It's a UX gate, not part of the wire format
  // — the picked entity ids are globally unique so SubscriptionMatch
  // doesn't need to round-trip an app reference.
  const [appId, setAppId] = useState<string | undefined>(undefined)

  const { org } = useOrg()
  const needsApp = kind === 'components' || kind === 'actions'

  // Apps list for the app picker. Loaded lazily — only when the user
  // navigates into the components/actions branch.
  const appsQuery = useQuery({
    queryKey: ['match-picker-apps', org.id],
    queryFn: () => getApps({ orgId: org.id, limit: 100 }),
    enabled: mode === 'specific' && needsApp,
  })
  const appOptions = useMemo(
    () =>
      (appsQuery.data?.data ?? []).map((a) => ({
        value: a.id ?? '',
        label: a.name ?? a.id ?? '',
      })),
    [appsQuery.data]
  )

  // Whenever any of the local controls change, project them back into the
  // SubscriptionMatch wire shape and notify the parent. This is the
  // single source-of-truth flow — the parent's `value` is purely
  // derived from these inputs while the picker is mounted.
  useEffect(() => {
    if (mode === 'all') {
      onChange(undefined)
      return
    }
    if (predicate === 'any') {
      onChange({ [kind]: {} })
      return
    }
    if (predicate === 'specific') {
      const ids = value?.[kind]?.ids ?? []
      onChange({ [kind]: { ids } })
      return
    }
    // labels
    const include = parseLabelsQuery(labelsRaw)
    const exclude = parseLabelsQuery(excludeLabelsRaw)
    // Mirror the Slack modal: emit only the populated halves so the
    // wire payload stays minimal. Empty include+exclude is invalid
    // (server-side Validate rejects it); we still emit the empty
    // shape so the submit-time validator can surface the error rather
    // than silently keeping the previous match.
    const selector: { match_labels?: typeof include; not_match_labels?: typeof exclude } = {}
    if (Object.keys(include).length > 0) selector.match_labels = include
    if (Object.keys(exclude).length > 0) selector.not_match_labels = exclude
    onChange({ [kind]: { selector } })
    // We intentionally exclude `value` and `onChange` from deps to avoid
    // an infinite loop — onChange is called with the derived value, which
    // the parent re-passes as `value`. Including either would re-fire the
    // effect on every parent render.
  }, [mode, kind, predicate, labelsRaw, excludeLabelsRaw])

  // Switching kind clears the entity ids to mirror the Slack modal's
  // clearStaleSpecificFields behaviour — ids belong to the previous
  // kind's taxonomy and would never resolve under the new one. The app
  // picker also resets because component/action ids are scoped to a
  // specific app.
  const handleKindChange = (next: TargetKind) => {
    if (next === kind) return
    setKind(next)
    setAppId(undefined)
    setLabelsRaw('')
    setExcludeLabelsRaw('')
  }

  // Switching the app inside components/actions invalidates the
  // currently selected ids — they belong to the previous app's
  // namespace. Clear them so the user starts the picker fresh.
  const handleAppChange = (nextAppId: string) => {
    if (nextAppId === appId) return
    setAppId(nextAppId)
    onChange({ [kind]: { ids: [] } })
  }

  return (
    <fieldset
      disabled={disabled}
      className={cn('flex flex-col gap-4', { 'opacity-60': disabled })}
    >
      <div className="flex flex-col gap-2">
        <Label>What should this match?</Label>
        <div className="flex flex-col gap-1">
          <RadioInput
            name="match-picker-mode"
            value="all"
            checked={mode === 'all'}
            onChange={() => setMode('all')}
            labelProps={{ labelText: 'Everything in this org' }}
            disabled={disabled}
          />
          <RadioInput
            name="match-picker-mode"
            value="specific"
            checked={mode === 'specific'}
            onChange={() => setMode('specific')}
            labelProps={{ labelText: 'Specific resources' }}
            disabled={disabled}
          />
        </div>
      </div>

      {mode === 'specific' ? (
        <>
          <div className="flex flex-col gap-2">
            <Label>Resource type</Label>
            <div className="flex flex-wrap gap-1">
              {ALL_TARGET_KINDS.map((k) => (
                <RadioInput
                  key={k}
                  name="match-picker-kind"
                  value={k}
                  checked={kind === k}
                  onChange={() => handleKindChange(k)}
                  labelProps={{ labelText: TARGET_KIND_LABELS[k] }}
                  disabled={disabled}
                />
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <Label>Match by</Label>
            <div className="flex flex-col gap-1">
              <RadioInput
                name="match-picker-predicate"
                value="any"
                checked={predicate === 'any'}
                onChange={() => setPredicate('any')}
                labelProps={{
                  labelText: `Any ${TARGET_KIND_LABELS_SINGULAR[kind]}`,
                }}
                disabled={disabled}
              />
              <RadioInput
                name="match-picker-predicate"
                value="specific"
                checked={predicate === 'specific'}
                onChange={() => setPredicate('specific')}
                labelProps={{
                  labelText: `Specific ${TARGET_KIND_LABELS_PLURAL[kind]}`,
                }}
                disabled={disabled}
              />
              <RadioInput
                name="match-picker-predicate"
                value="labels"
                checked={predicate === 'labels'}
                onChange={() => setPredicate('labels')}
                labelProps={{ labelText: 'By labels' }}
                disabled={disabled}
              />
            </div>
          </div>

          {predicate === 'specific' && needsApp ? (
            <div className="flex flex-col gap-2">
              <Select
                labelProps={{ labelText: 'App' }}
                placeholder={
                  appsQuery.isLoading ? 'Loading apps…' : 'Select an app…'
                }
                options={appOptions}
                value={appId ?? ''}
                onChange={(e) => handleAppChange(e.target.value)}
                searchable
                disabled={disabled || appsQuery.isLoading}
              />
              <Text variant="subtext" theme="neutral">
                {TARGET_KIND_LABELS[kind]} are scoped to a single app. To match
                across apps, use <em>By labels</em> instead.
              </Text>
            </div>
          ) : null}

          {predicate === 'specific' && (!needsApp || appId) ? (
            <div className="flex flex-col gap-2">
              <Label>{TARGET_KIND_LABELS[kind]}</Label>
              <EntityMultiSelect
                kind={kind}
                appId={appId}
                selectedIds={value?.[kind]?.ids ?? []}
                onChange={(ids) => onChange({ [kind]: { ids } })}
                disabled={disabled}
              />
            </div>
          ) : null}

          {predicate === 'labels' ? (
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-2">
                <Input
                  labelProps={{ labelText: 'Include labels' }}
                  placeholder="env=prod, tier=critical, owner=*"
                  value={labelsRaw}
                  onChange={(e) => setLabelsRaw(e.target.value)}
                  disabled={disabled}
                />
                <Text variant="subtext" theme="neutral">
                  Match {TARGET_KIND_LABELS_PLURAL[kind]} where every listed key
                  matches. Use <code>key=*</code> (or just <code>key</code>) to
                  match any value for a key. Leave blank to match everything
                  except the exclusions below.
                </Text>
              </div>
              <div className="flex flex-col gap-2">
                <Input
                  labelProps={{ labelText: 'Exclude labels' }}
                  placeholder="env=stage"
                  value={excludeLabelsRaw}
                  onChange={(e) => setExcludeLabelsRaw(e.target.value)}
                  disabled={disabled}
                />
                <Text variant="subtext" theme="neutral">
                  Drop {TARGET_KIND_LABELS_PLURAL[kind]} where any listed key
                  matches — e.g. <code>env=stage</code> to skip stage installs
                  without enumerating every prod env value.
                </Text>
              </div>
              {labelsRaw.trim().length === 0 &&
              excludeLabelsRaw.trim().length === 0 ? (
                <Banner theme="warn">
                  Add at least one include or exclude key (e.g.{' '}
                  <code>env=prod</code> or <code>env=stage</code>) — an empty
                  selector matches nothing.
                </Banner>
              ) : null}
            </div>
          ) : null}
        </>
      ) : null}
    </fieldset>
  )
}

// derivePredicate inspects an existing SubscriptionMatch to figure out
// which predicate radio should start out checked. Mirrors the slack
// modal's render-state derivation: selector → labels, ids → specific,
// otherwise any (which covers both the empty TargetMatch{} and an
// undefined match for the chosen kind).
const derivePredicate = (
  m: SubscriptionMatch | undefined,
  kind: TargetKind
): Predicate => {
  const tm = m?.[kind]
  if (!tm) return 'any'
  if (tm.selector?.match_labels || tm.selector?.not_match_labels) return 'labels'
  if (tm.ids && tm.ids.length > 0) return 'specific'
  return 'any'
}
