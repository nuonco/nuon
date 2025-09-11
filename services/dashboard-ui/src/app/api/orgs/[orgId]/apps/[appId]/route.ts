import { type NextRequest, NextResponse } from "next/server";
import { getAppById } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<"orgId" | "appId">,
) {
  const { appId, orgId } = await params;
  const response = await getAppById({ appId, orgId });
  return NextResponse.json(response);
}
