'use server'

import { revalidatePath } from 'next/cache'
import type { ICreateInstallData } from '@/lib'
import type { TBuild, TComponent } from '@/types'
import { API_URL, nueMutateData, getFetchOpts } from '@/utils'

interface IRevalidateAppData {
  appId: string
  orgId: string
}

export async function revalidateAppData({ appId, orgId }: IRevalidateAppData) {
  revalidatePath(`/${orgId}/apps/${appId}`)
}

interface ICreateComponentBuild {
  componentId: string
  orgId: string
}

export async function createComponentBuild({
  componentId,
  orgId,
}: ICreateComponentBuild) {
  return nueMutateData<TBuild>({
    path: `components/${componentId}/builds`,
    orgId,
    body: { use_latest: true },
  })
}

interface ICreateAppInstall {
  appId: string
  orgId: string
  formData: FormData
  platform: string | 'aws' | 'azure'
}

export async function createAppInstall({
  appId,
  orgId,
  formData: fd,
  platform,
}: ICreateAppInstall) {
  const formData = Object.fromEntries(fd)

  const inputs = Object.keys(formData).reduce((acc, key) => {
    if (key.includes('inputs:')) {
      let value: any = formData[key]
      if (value === 'on' || value === 'off') {
        value = Boolean(value === 'on').toString()
      }

      acc[key.replace('inputs:', '')] = value
    }

    return acc
  }, {})

  let data: ICreateInstallData = {
    inputs,
    name: formData.name as string,
  }

  if (platform === 'aws') {
    data = {
      aws_account: {
        // iam_role_arn: formData?.iam_role_arn as string,
        iam_role_arn: 'old-field',
        region: formData?.region as string,
      },
      ...data,
    }
  }

  if (platform === 'azure') {
    data = {
      azure_account: {
        location: formData?.location as string,
        service_principal_app_id: formData?.service_principal_app_id as string,
        service_principal_password:
          formData?.service_principal_password as string,
        subscription_id: formData?.subscription_id as string,
        subscription_tenant_id: formData?.subscription_tenant_id as string,
      },
      ...data,
    }
  }

  const res = fetch(`${API_URL}/v1/apps/${appId}/installs`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify(data),
    method: 'POST',
  })
    .then(async (r) => {
      if (!r.ok) {
        throw new Error('Unable to create inputs')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  const response = await res
  const workflowId = response.headers.get('x-nuon-install-workflow-id')
  const install = await response.json()

  return {
    installId: install?.id,
    workflowId,
  }
}

interface IBuildComponents {
  appId: string
  components: Array<TComponent>
  orgId: string
}

export async function buildComponents({
  appId,
  components,
  orgId,
}: IBuildComponents) {
  return Promise.all(
    components.map(
      async ({ id, name }) =>
        await nueMutateData<TBuild>({
          path: `components/${id}/builds`,
          orgId,
          body: { use_latest: true },
        }).then((res) =>
          res?.error
            ? {
                ...res,
                error: {
                  ...res?.error,
                  meta: { name, id },
                },
              }
            : res
        )
    )
  ).then((res) => {
    revalidatePath(`/${orgId}/apps/${appId}/components`)
    return res
  })
}
