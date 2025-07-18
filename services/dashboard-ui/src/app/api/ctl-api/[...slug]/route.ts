import { NextRequest, NextResponse } from "next/server";
import { API_URL } from "@/utils/configs";

function buildTargetUrl(params: { slug?: string[] }, req: NextRequest) {
  const urlPath = params.slug && params.slug.length > 0 ? "/" + params.slug.join("/") : "";
  return `${API_URL}${urlPath}${req.nextUrl.search}`;
}

async function proxyRequest(
  method: string,
  req: NextRequest,
  params: { slug?: string[] }
) {
  const targetUrl = buildTargetUrl(params, req);

  let fetchInit: RequestInit = {
    method,
    headers: { ...Object.fromEntries(req.headers) },
  };

  if (["POST", "PUT", "PATCH", "DELETE"].includes(method)) {
    const contentType = req.headers.get("content-type") || "";
    if (contentType.includes("application/json")) {
      fetchInit.body = JSON.stringify(await req.json());
    } else {
      fetchInit.body = await req.text();
    }
  }

  const res = await fetch(targetUrl, fetchInit);
  const body = await res.text();

  return new NextResponse(body, {
    status: res.status,
    headers: res.headers,
  });
}

// No explicit type for `context`!
export async function GET(req: NextRequest, context: any) {
  return proxyRequest("GET", req, context.params);
}
export async function POST(req: NextRequest, context: any) {
  return proxyRequest("POST", req, context.params);
}
export async function PUT(req: NextRequest, context: any) {
  return proxyRequest("PUT", req, context.params);
}
export async function DELETE(req: NextRequest, context: any) {
  return proxyRequest("DELETE", req, context.params);
}
export async function PATCH(req: NextRequest, context: any) {
  return proxyRequest("PATCH", req, context.params);
}
export async function OPTIONS(req: NextRequest, context: any) {
  return proxyRequest("OPTIONS", req, context.params);
}
export async function HEAD(req: NextRequest, context: any) {
  return proxyRequest("HEAD", req, context.params);
}
