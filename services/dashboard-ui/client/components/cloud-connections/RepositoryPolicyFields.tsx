import type { AnyFieldApi } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { fieldErrorMessage } from '@/components/common/form/field-error'
import type { TCloudConnection } from '@/types'

export const RepositoryPolicyFields = ({
  field,
  id,
  preset,
  disabled = false,
}: {
  field: AnyFieldApi
  id: string
  preset: TCloudConnection['preset']
  disabled?: boolean
}) => {
  const names = ((field.state.value as string | undefined) || '').split(/\r?\n/)
  const errorMessage = fieldErrorMessage(field)
  const rowLabel =
    preset === 'gar'
      ? 'GAR repository ID'
      : preset === 'acr'
        ? 'ACR repository path'
        : 'ECR repository name'

  return (
    <fieldset disabled={disabled} className="flex min-w-0 flex-col gap-3">
      <legend className="mb-2">
        <Text weight="strong">
          {preset === 'gar' ? 'Repository IDs' : 'Repository names'}
        </Text>
      </legend>
      {names.map((name, index) => (
        <div key={index} className="flex items-center gap-2">
          <div className="min-w-0 flex-1">
            <Input
              id={`${id}-${index}`}
              name={`${field.name}-${index}`}
              aria-label={`${rowLabel} ${index + 1}`}
              aria-describedby={`${id}-description`}
              value={name}
              onChange={(event) =>
                field.handleChange(
                  names
                    .map((value, i) =>
                      i === index ? event.target.value : value
                    )
                    .join('\n')
                )
              }
              onBlur={field.handleBlur}
              placeholder={preset === 'gar' ? 'acme-images' : 'acme/api'}
              className="!font-mono"
              spellCheck={false}
              autoCapitalize="none"
              autoCorrect="off"
              disabled={disabled}
              error={!!errorMessage}
            />
          </div>
          <Button
            type="button"
            variant="icon"
            aria-label={`Remove ${rowLabel} ${index + 1}`}
            disabled={disabled}
            onClick={() => {
              field.handleChange(names.filter((_, i) => i !== index).join('\n'))
              requestAnimationFrame(() =>
                document
                  .getElementById(`${id}-${Math.max(0, index - 1)}`)
                  ?.focus()
              )
            }}
          >
            <Icon variant="TrashIcon" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        className="w-fit"
        disabled={disabled}
        onClick={() => {
          field.handleChange([...names, ''].join('\n'))
          requestAnimationFrame(() =>
            document.getElementById(`${id}-${names.length}`)?.focus()
          )
        }}
      >
        <Icon variant="PlusIcon" />
        Add repository
      </Button>
      <Text
        id={`${id}-description`}
        variant="subtext"
        theme={errorMessage ? 'error' : 'neutral'}
        aria-live="polite"
      >
        {errorMessage ||
          (preset === 'gar'
            ? 'Use repository IDs, not image paths or tags. A GAR repository can contain multiple images.'
            : 'Use exact repository names, including nested paths. No registry hosts, wildcards, tags, or digests.')}
      </Text>
    </fieldset>
  )
}
