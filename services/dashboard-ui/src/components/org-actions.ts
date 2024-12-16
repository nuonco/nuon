"use server"

import { postJoinWaitlist } from "@/lib"

export async function requestWaitlistAccess(formData: FormData): Promise<Record<"org_name"| string, string>> {
  const data = Object.fromEntries(formData);

  return postJoinWaitlist(data);
}

