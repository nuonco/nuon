import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { RepositoryPolicyFields } from './RepositoryPolicyFields'
import { repositoryNamesSchema } from './repository-scope'

export default { title: 'Features / Cloud connections / Policy inputs' }

const Demo = ({
  preset = 'ecr',
  repositories = 'acme/api\nacme/platform/worker',
  disabled = false,
}: {
  preset?: 'ecr' | 'gar' | 'acr'
  repositories?: string
  disabled?: boolean
}) => {
  const form = useForm({
    defaultValues: { repositories },
    validators: {
      onChange: z.object({
        repositories: repositoryNamesSchema(preset),
      }),
    },
  })
  return (
    <div className="max-w-2xl p-4">
      <form.Field name="repositories">
        {(field) => (
          <RepositoryPolicyFields
            field={field}
            id="repository-policy"
            preset={preset}
            disabled={disabled}
          />
        )}
      </form.Field>
    </div>
  )
}

export const ECR = () => <Demo />
export const GAR = () => (
  <Demo preset="gar" repositories="acme-images\nacme-services" />
)
export const ACR = () => <Demo preset="acr" />
export const Empty = () => <Demo repositories="" />
export const Disabled = () => <Demo disabled />
