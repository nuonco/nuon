import { type NextRequest, NextResponse } from "next/server";
import { getApp } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<"orgId" | "appId">,
) {
  const { appId, orgId } = await params;
  const response = await getApp({ appId, orgId });
  return NextResponse.json(response);
}
