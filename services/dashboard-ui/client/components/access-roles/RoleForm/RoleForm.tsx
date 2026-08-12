import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormTextarea } from '@/components/common/form/FormTextarea'
import { Label } from '@/components/common/form/Label'
import { EntityMultiSelect } from '@/components/match'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TRoleContext } from '@/types'
import { entriesValid, readAllWriteScoped } from '../permissions'
import { FormPermissionEntries } from '../PermissionEntries'
import { roleFormSchema, type RoleFormValues } from './schema'

const CONTEXT_OPTIONS: { value: TRoleContext; label: string }[] = [
  { value: 'team', label: 'Team members' },
  { value: 'service_account', label: 'Service accounts' },
  { value: 'api_token', label: 'API tokens' },
  { value: 'oidc_trust_policy', label: 'OIDC trust policies' },
]

export const RoleFormModal = ({
  mode,
  orgId,
  initialValues,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  mode: 'create' | 'edit'
  orgId: string
  initialValues?: Partial<RoleFormValues>
  isPending: boolean
  error: TAPIError | null
  onSubmit: (values: RoleFormValues) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: {
      title: initialValues?.title ?? '',
      description: initialValues?.description ?? '',
      contexts:
        initialValues?.contexts ??
        ['team', 'service_account', 'api_token', 'oidc_trust_policy'],
      permissions: initialValues?.permissions ?? [],
    } as RoleFormValues,
    validators: {
      onMount: roleFormSchema,
      onChange: roleFormSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const permissions = useStore(form.store, (s) => s.values.permissions)
  const contexts = useStore(form.store, (s) => s.values.contexts)
  const permissionsValid = entriesValid(permissions)

  return (
    <Modal
      size="xl"
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ShieldCheckIcon" size="24" />
          {mode === 'create' ? 'Create role' : 'Edit role'}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending
          ? mode === 'create'
            ? 'Creating role'
            : 'Saving role'
          : mode === 'create'
            ? 'Create role'
            : 'Save role',
        disabled: !canSubmit || !permissionsValid || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner
          error={error}
          fallback={
            mode === 'create' ? 'Unable to create role' : 'Unable to save role'
          }
        />

        <form.Field name="title">
          {(field) => (
            <FormInput
              field={field}
              id="role-title"
              placeholder="e.g. Release manager"
              type="text"
              disabled={isPending}
              labelProps={{ labelText: 'Name' }}
            />
          )}
        </form.Field>

        <form.Field name="description">
          {(field) => (
            <FormTextarea
              field={field}
              id="role-description"
              rows={2}
              placeholder="What this role is for"
              disabled={isPending}
              labelProps={{ labelText: 'Description' }}
            />
          )}
        </form.Field>

        <form.Field name="contexts">
          {(field) => (
            <div className="flex flex-col gap-1">
              <Label htmlFor="role-contexts">
                <Text variant="body" className="font-medium">
                  Can be assigned to
                </Text>
              </Label>
              <div
                id="role-contexts"
                className="flex flex-wrap items-center gap-1"
              >
                {CONTEXT_OPTIONS.map((option) => (
                  <CheckboxInput
                    key={option.value}
                    checked={(field.state.value ?? []).includes(option.value)}
                    disabled={isPending}
                    onChange={(e) =>
                      field.handleChange(
                        e.target.checked
                          ? [...(field.state.value ?? []), option.value]
                          : (field.state.value ?? []).filter(
                              (c: TRoleContext) => c !== option.value
                            )
                      )
                    }
                    labelProps={{
                      labelText: option.label,
                      labelTextProps: { variant: 'subtext' },
                    }}
                  />
                ))}
              </div>
              {(contexts ?? []).length === 0 ? (
                <Text variant="subtext" theme="warn">
                  With nothing selected the role exists but no picker offers it.
                </Text>
              ) : null}
            </div>
          )}
        </form.Field>

        {mode === 'create' && permissions.length === 0 ? (
          <ReadAllWriteScopedPreset
            orgId={orgId}
            disabled={isPending}
            onApply={(entries) =>
              form.setFieldValue('permissions', entries)
            }
          />
        ) : null}

        <form.Field name="permissions">
          {(field) => (
            <FormPermissionEntries field={field} disabled={isPending} />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}

// The read-all/write-scoped shape is the one custom roles are usually reaching
// for, and it is easy to get subtly wrong by hand: without the org read entry
// the role can open an install but cannot list installs, because collection
// endpoints authorize at the org tier.
const ReadAllWriteScopedPreset = ({
  orgId,
  disabled,
  onApply,
}: {
  orgId: string
  disabled?: boolean
  onApply: (entries: ReturnType<typeof readAllWriteScoped>) => void
}) => {
  const [installIds, setInstallIds] = useState<string[]>([])

  return (
    <div className="flex flex-col gap-3 rounded-md border p-4">
      <div className="flex flex-col gap-1">
        <Text variant="body" weight="strong">
          Start from a preset
        </Text>
        <Text variant="subtext" theme="neutral">
          Read access to everything in the org, write access only to the
          installs you pick.
        </Text>
      </div>

      <EntityMultiSelect
        kind="installs"
        selectedIds={installIds}
        onChange={setInstallIds}
        disabled={disabled}
      />

      <div>
        <Button
          variant="secondary"
          size="sm"
          disabled={disabled || installIds.length === 0}
          onClick={() => onApply(readAllWriteScoped({ orgId, installIds }))}
        >
          <Icon variant="SparkleIcon" size={14} />
          Apply preset
        </Button>
      </div>

      <Banner theme="info">
        You can edit or remove any of the permissions it adds.
      </Banner>
    </div>
  )
}
