import { api } from '@/lib/api'

export const postPhoneHome = ({
  installId,
  phoneHomeId,
  body,
}: {
  installId: string
  phoneHomeId: string
  body: Record<string, unknown>
}) =>
  api<string>({
    method: 'POST',
    path: `installs/${installId}/phone-home/${phoneHomeId}`,
  body,
  })
