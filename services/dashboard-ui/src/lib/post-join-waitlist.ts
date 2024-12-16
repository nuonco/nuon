import { API_URL, getFetchOpts } from '@/utils'

export async function postJoinWaitlist(data: Record<string, FormDataEntryValue>): Promise<Record<"org_name" | string, string>> {
  const res = await fetch(`${API_URL}/v1/general/waitlist`, {
    ...(await getFetchOpts()),
    body: JSON.stringify(data),
    method: 'POST',
  })

  if (!res.ok) {
    throw new Error("Couldn't add you to the waitlist, try again.")
  }

  return res.json()
}
