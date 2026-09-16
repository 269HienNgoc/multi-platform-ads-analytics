const backendURL = (process.env.BACKEND_API_URL ?? "http://127.0.0.1:8080").replace(/\/$/, "");

type RouteParameters = {
  params: Promise<{ path: string[] }>;
};

async function forward(request: Request, context: RouteParameters) {
  const { path } = await context.params;
  if (path.length === 0 || path.some((segment) => segment === "." || segment === "..")) {
    return Response.json(
      { error: { code: "invalid_proxy_path", message: "Đường dẫn backend không hợp lệ." } },
      { status: 400 },
    );
  }

  const incomingURL = new URL(request.url);
  const targetURL = new URL(`${backendURL}/${path.map(encodeURIComponent).join("/")}`);
  targetURL.search = incomingURL.search;

  const headers = new Headers({ Accept: "application/json" });
  const contentType = request.headers.get("content-type");
  const requestID = request.headers.get("x-request-id");
  const idempotencyKey = request.headers.get("idempotency-key");
  if (contentType) headers.set("Content-Type", contentType);
  if (requestID) headers.set("X-Request-ID", requestID);
  if (idempotencyKey) headers.set("Idempotency-Key", idempotencyKey);
  if (process.env.BACKEND_API_KEY) headers.set("X-API-Key", process.env.BACKEND_API_KEY);

  const hasBody = request.method !== "GET" && request.method !== "HEAD";
  try {
    const proxyRequest = new Request(targetURL, {
      method: request.method,
      headers,
      body: hasBody ? request.body : undefined,
      redirect: "manual",
      signal: request.signal,
      ...(hasBody ? { duplex: "half" } : {}),
    } as RequestInit & { duplex?: "half" });
    const upstream = await fetch(proxyRequest, { cache: "no-store" });
    const responseHeaders = new Headers({
      "Cache-Control": "no-store",
      "Content-Type": upstream.headers.get("content-type") ?? "application/json",
    });
    const upstreamRequestID = upstream.headers.get("x-request-id");
    if (upstreamRequestID) responseHeaders.set("X-Request-ID", upstreamRequestID);

    return new Response(upstream.body, {
      status: upstream.status,
      headers: responseHeaders,
    });
  } catch {
    return Response.json(
      { error: { code: "backend_unavailable", message: "Không thể kết nối tới backend." } },
      { status: 502, headers: { "Cache-Control": "no-store" } },
    );
  }
}

export const GET = forward;
export const POST = forward;
export const PUT = forward;
export const PATCH = forward;
export const DELETE = forward;
