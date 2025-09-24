'use server'

import { revalidatePath } from 'next/cache'
import { createAppInstall as create, type TCreateAppInstallBody } from '@/lib'

export async function createAppInstall({
  appId,
  formData: fd,
  orgId,
  path,
}: {
  appId: string
  formData: FormData
  orgId: string
  path: string
}) {
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

  let body: TCreateAppInstallBody = {
    inputs,
    name: formData?.name as string,
    metadata: {
      managed_by: 'nuon/dashboard',
    },
  }

  if (formData?.region) {
    body = {
      ...body,
      aws_account: {
        iam_role_arn: '',
        region: formData?.region as string,
      },
    }
  }

  if (formData?.location) {
    body = {
      ...body,
      azure_accout: {
        location: formData?.location as string,
        service_principal_app_id: '',
        service_principal_password: '',
        subscription_id: '',
        subscription_tenant_id: '',
      },
    }
  }

  return create({ appId, orgId, body }).then((res) => {
    revalidatePath(path)
    return res
  })
}
