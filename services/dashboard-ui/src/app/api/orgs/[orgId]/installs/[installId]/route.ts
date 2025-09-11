import { type NextRequest, NextResponse } from "next/server";
import { getInstallById } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<"orgId" | "installId">,
) {
  const { installId, orgId } = await params;
  const response = await getInstallById({ installId, orgId });
  return NextResponse.json(response);
}
