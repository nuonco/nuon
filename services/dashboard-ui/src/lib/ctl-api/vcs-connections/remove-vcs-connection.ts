import { api } from "@/lib/api";

export async function removeVCSConnection({ orgId, connectionId }: { orgId: string, connectionId: string }) {
  return api({
    method: "DELETE",
    orgId,
    path: `vcs/connections/${connectionId}`,   
  })
}
