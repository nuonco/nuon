import { api } from '@/lib/api'

export const postPhoneHome = ({
  installId,
  orgId,
  phoneHomeId,
  body,
}: {
  installId: string
  orgId: string
  phoneHomeId: string
  body: Record<string, unknown>
}) =>
  api<string>({
    method: 'POST',
    path: `installs/${installId}/phone-home/${phoneHomeId}`,
    orgId,
    body,
  })
