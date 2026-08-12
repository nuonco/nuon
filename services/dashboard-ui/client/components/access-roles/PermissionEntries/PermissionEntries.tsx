import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import type { TPermissionEntry, TPermissionResourceType } from '@/types'
import {
  RESOURCE_TYPE_OPTIONS,
  canScopeWildcard,
  entryError,
  isWildcard,
  newEntry,
} from '../permissions'
import { VerbsDropdown } from './VerbsDropdown'

export interface IPermissionEntries {
  value: TPermissionEntry[]
  onChange: (next: TPermissionEntry[]) => void
  appOptions: { value: string; label: string }[]
  installOptions: { value: string; label: string }[]
  branchOptions: { value: string; label: string }[]
  branchesLoading?: boolean
  orgId: string
  disabled?: boolean
}

const TARGET_ALL = '__all__'

const RESOURCE_TYPE_PLURALS: Record<TPermissionResourceType, string> = {
  install: 'installs',
  app: 'apps',
  app_branch: 'branches',
  org: 'org',
}

export const PermissionEntries = ({
  value,
  onChange,
  appOptions,
  installOptions,
  branchOptions,
  branchesLoading,
  orgId,
  disabled,
}: IPermissionEntries) => {
  const update = (index: number, entry: TPermissionEntry) =>
    onChange(value.map((e, i) => (i === index ? entry : e)))

  const remove = (index: number) =>
    onChange(value.filter((_, i) => i !== index))

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <Label htmlFor="permission-entries">
          <Text variant="body" className="font-medium">
            Permissions
          </Text>
        </Label>
        <Button
          variant="secondary"
          size="xs"
          disabled={disabled}
          onClick={() => onChange([...value, newEntry()])}
        >
          <Icon variant="PlusIcon" size={14} />
          Add permission
        </Button>
      </div>

      {value.length === 0 ? (
        <div className="rounded-md border border-dashed p-6 text-center">
          <Text variant="subtext" theme="neutral">
            No permissions yet. A role needs at least one.
          </Text>
        </div>
      ) : (
        <div id="permission-entries" className="flex flex-col gap-3">
          {value.map((entry, index) => (
            <EntryRow
              key={index}
              index={index}
              entry={entry}
              appOptions={appOptions}
              installOptions={installOptions}
              branchOptions={branchOptions}
              branchesLoading={branchesLoading}
              orgId={orgId}
              disabled={disabled}
              onChange={(next) => update(index, next)}
              onRemove={() => remove(index)}
            />
          ))}
        </div>
      )}
    </div>
  )
}

const EntryRow = ({
  index,
  entry,
  appOptions,
  installOptions,
  branchOptions,
  branchesLoading,
  orgId,
  disabled,
  onChange,
  onRemove,
}: {
  index: number
  entry: TPermissionEntry
  appOptions: { value: string; label: string }[]
  installOptions: { value: string; label: string }[]
  branchOptions: { value: string; label: string }[]
  branchesLoading?: boolean
  orgId: string
  disabled?: boolean
  onChange: (next: TPermissionEntry) => void
  onRemove: () => void
}) => {
  const error = entryError(entry)
  const wildcard = isWildcard(entry)

  const setResourceType = (raw: string) => {
    const resourceType = raw as TPermissionResourceType
    onChange({
      resource_type: resourceType,
      resource_id: resourceType === 'org' ? orgId : '*',
      permissions: entry.permissions,
    })
  }

  // Picking a specific resource has to drop any scope, since the API only
  // accepts a scope on wildcard entries.
  const setTarget = (raw: string) => {
    onChange({
      resource_type: entry.resource_type,
      resource_id: raw === TARGET_ALL ? '*' : raw,
      permissions: entry.permissions,
    })
  }

  const plural = RESOURCE_TYPE_PLURALS[entry.resource_type]

  const resourceOptions =
    entry.resource_type === 'app'
      ? appOptions
      : entry.resource_type === 'app_branch'
        ? branchOptions
        : installOptions

  const targetOptions = [
    { value: TARGET_ALL, label: `All ${plural}` },
    ...resourceOptions,
  ]

  const targetValue = wildcard ? TARGET_ALL : entry.resource_id

  // An id set through the API, or one past the page of resources we fetched,
  // has no matching option. Without an entry for it the select would render
  // empty and the next change would silently rewrite the target.
  if (
    targetValue &&
    !targetOptions.some((option) => option.value === targetValue)
  ) {
    targetOptions.push({ value: targetValue, label: targetValue })
  }

  return (
    <div className="flex flex-col gap-3 rounded-md border p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="flex flex-1 flex-wrap items-start gap-3">
          <div className="min-w-40 flex-1">
            <VerbsDropdown
              id={`permission-actions-${index}`}
              value={entry.permissions}
              onChange={(permissions) => onChange({ ...entry, permissions })}
              disabled={disabled}
            />
          </div>

          <div className="min-w-40 flex-1">
            <Select
              options={RESOURCE_TYPE_OPTIONS}
              value={entry.resource_type}
              onChange={setResourceType}
              disabled={disabled}
              size="sm"
              labelProps={{ labelText: 'On' }}
              helperText={
                entry.resource_type === 'org'
                  ? 'Org-level resources, including the lists every page reads.'
                  : undefined
              }
            />
          </div>

          {entry.resource_type === 'org' ? null : (
            <div className="min-w-40 flex-1">
              <Select
                options={targetOptions}
                value={targetValue}
                onChange={setTarget}
                disabled={disabled}
                searchable
                size="sm"
                placeholder={
                  branchesLoading && entry.resource_type === 'app_branch'
                    ? 'Loading branches'
                    : `Search ${plural}`
                }
                labelProps={{ labelText: 'Which' }}
              />
            </div>
          )}

          {wildcard && canScopeWildcard(entry.resource_type) ? (
            <div className="min-w-40 flex-1">
              <Select
                options={[{ value: '', label: 'In any app' }, ...appOptions]}
                value={entry.scope_id ?? ''}
                onChange={(scopeID) =>
                  onChange({
                    ...entry,
                    scope_type: scopeID ? 'app' : undefined,
                    scope_id: scopeID || undefined,
                  })
                }
                disabled={disabled}
                searchable
                size="sm"
                labelProps={{ labelText: 'Scoped to' }}
              />
            </div>
          ) : null}
        </div>

        <Button
          variant="ghost"
          size="xs"
          className="!p-2 !text-red-800 dark:!text-red-500"
          disabled={disabled}
          onClick={onRemove}
          aria-label="Remove permission"
        >
          <Icon variant="TrashIcon" size={16} />
        </Button>
      </div>

      {error ? (
        <Badge theme="error" size="sm">
          {error}
        </Badge>
      ) : null}
    </div>
  )
}
