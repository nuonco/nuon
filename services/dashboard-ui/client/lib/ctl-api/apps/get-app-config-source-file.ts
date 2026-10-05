import { api } from '@/lib/api'

export type TAppConfigSourceFile = {
  content: string
  filename: string
}

export const getAppConfigSourceFile = ({
  appId,
  configId,
  path,
  orgId,
}: {
  appId: string
  configId: string
  path: string
  orgId: string
}) =>
  api<TAppConfigSourceFile>({
    path: `apps/${appId}/configs/${configId}/source-files/${path}`,
    orgId,
  })
