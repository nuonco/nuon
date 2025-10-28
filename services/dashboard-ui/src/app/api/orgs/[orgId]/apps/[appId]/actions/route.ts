import { type NextRequest, NextResponse } from "next/server";
import { getActions } from "@/lib";
import type { TRouteProps } from "@/types";

export async function GET(request: NextRequest, { params }: TRouteProps<"orgId" | "appId">) {
  const { orgId, appId } = await params;
  const { searchParams } = new URL(request.url);
  const limit = searchParams.get("limit") || undefined;
  const offset = searchParams.get("offset") || undefined;
  const q = searchParams.get("q") || undefined;
  
  const response = await getActions({ appId, orgId, limit, offset, q });
  return NextResponse.json(response);
}
