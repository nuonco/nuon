"use server";

import { revalidatePath } from "next/cache";
import { removeVCSConnection as removeVCS } from "@/lib";

export async function removeVCSConnection({
  orgId,
  path,
  connectionId,
}: {
  orgId: string;
  path?: string;
  connectionId: string;
}) {
  return removeVCS({ orgId, connectionId }).then((r) => {
    if (path) revalidatePath(path);
    return r;
  });
}
