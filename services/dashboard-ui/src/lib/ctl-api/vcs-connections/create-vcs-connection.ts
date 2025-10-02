import { api } from '@/lib/api'

type TCreateVCSConnectionBody = {
  github_install_id: string
}

export async function createVCSConnection({
  body,
  orgId,
}: {
  body: TCreateVCSConnectionBody
  orgId: string
}) {
  return api({
    method: 'POST',
    orgId,
    path: `vcs/connections`,
    body,
  })
}
