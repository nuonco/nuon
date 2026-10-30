import { useState } from 'react'
import type { TCloudConnection } from '@/types'
import { AddRepositoryAccess } from '../AddRepositoryAccess/AddRepositoryAccess'
import { canonicalRepositories } from '../repository-scope'

const policy = (
  connection: TCloudConnection,
  names: string[]
): Record<string, unknown> =>
  connection.preset === 'gar'
    ? {
        bindings: canonicalRepositories(connection, names).map((resource) => ({
          resource,
          role: 'roles/artifactregistry.reader',
          member: `serviceAccount:${connection.principal}`,
        })),
      }
    : connection.preset === 'acr'
      ? {
          scope: connection.azure_registry_id,
          role: 'Container Registry Repository Reader',
          repositories: names,
          condition_version: '2.0',
        }
      : {
          Version: '2012-10-17',
          Statement: [
            {
              Effect: 'Allow',
              Action: 'ecr:GetAuthorizationToken',
              Resource: '*',
            },
            {
              Effect: 'Allow',
              Action: ['ecr:BatchGetImage', 'ecr:GetDownloadUrlForLayer'],
              Resource: canonicalRepositories(connection, names),
            },
          ],
        }

const commands = (connection: TCloudConnection, names: string[]) => {
  if (connection.preset === 'gar')
    return names
      .map(
        (name) =>
          `gcloud artifacts repositories add-iam-policy-binding '${name}' --project='acme-project' --location='us-central1' --role='roles/artifactregistry.reader' --member='serviceAccount:nuon-images@acme-project.iam.gserviceaccount.com'`
      )
      .join('\n')
  if (connection.preset === 'acr') {
    const condition = `((!(ActionMatches{'Microsoft.ContainerRegistry/registries/repositories/content/read'}) AND !(ActionMatches{'Microsoft.ContainerRegistry/registries/repositories/metadata/read'})) OR (${names.map((name) => `@Request[Microsoft.ContainerRegistry/registries/repositories:name] StringEqualsIgnoreCase '${name}'`).join(' OR ')}))`
    return `SERVICE_PRINCIPAL_ID=$(az ad sp show --id '${connection.principal}' --query id --output tsv)\naz role assignment create --assignee-object-id "$SERVICE_PRINCIPAL_ID" --assignee-principal-type ServicePrincipal --role 'Container Registry Repository Reader' --scope '${connection.azure_registry_id}' --condition-version '2.0' --condition "${condition}"`
  }
  return `aws iam put-role-policy --role-name acme-nuon --policy-name nuon-ecr-acme-example --policy-document '${JSON.stringify(
    {
      Version: '2012-10-17',
      Statement: [
        { Effect: 'Allow', Action: 'ecr:GetAuthorizationToken', Resource: '*' },
        {
          Effect: 'Allow',
          Action: ['ecr:BatchGetImage', 'ecr:GetDownloadUrlForLayer'],
          Resource: names.map(
            (name) => `arn:aws:ecr:us-west-2:123456789012:repository/${name}`
          ),
        },
      ],
    }
  )}'`
}

export const RepositoryAccessDemo = ({
  connection,
  generated = false,
  state = 'ready',
}: {
  connection: TCloudConnection
  generated?: boolean
  state?: 'ready' | 'generating' | 'error'
}) => {
  const names =
    connection.preset === 'gar'
      ? ['acme-images', 'acme-services']
      : ['acme/api', 'acme/worker']
  const [result, setResult] = useState(
    generated
      ? {
          names,
          permissionsCli: commands(connection, names),
          permissionsPolicy: policy(connection, names),
        }
      : undefined
  )
  const [error, setError] = useState(
    state === 'error'
      ? {
          error: 'Command generation failed',
          description: 'Try generating the commands again.',
          user_error: true,
        }
      : null
  )
  return (
    <AddRepositoryAccess
      connection={connection}
      repositories={generated || state !== 'ready' ? names.join('\n') : ''}
      isPending={state === 'generating'}
      error={error}
      onGenerate={(names) => {
        setError(null)
        setResult({
          names,
          permissionsCli: commands(connection, names),
          permissionsPolicy: policy(connection, names),
        })
      }}
      result={state === 'generating' || error ? undefined : result}
    />
  )
}
