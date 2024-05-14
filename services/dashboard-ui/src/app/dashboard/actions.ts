"use server";

import { redirect } from "next/navigation"
import { postOrg } from "@/lib"

export async function createOrg(formData: FormData) {
  const data = Object.fromEntries(formData)
  const org  = await postOrg(data as Record<string, string>)

  redirect(`/dashboard/${org?.id}`)  
}
