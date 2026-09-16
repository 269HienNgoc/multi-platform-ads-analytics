import { type NextRequest, NextResponse } from "next/server";

function unauthorized() {
  return new NextResponse("Authentication required", {
    status: 401,
    headers: { "WWW-Authenticate": 'Basic realm="AdPilot", charset="UTF-8"' },
  });
}

async function digest(value: string) {
  return new Uint8Array(
    await crypto.subtle.digest("SHA-256", new TextEncoder().encode(value)),
  );
}

async function credentialsMatch(actual: string, expected: string) {
  const [actualHash, expectedHash] = await Promise.all([digest(actual), digest(expected)]);
  let difference = actualHash.length ^ expectedHash.length;
  for (let index = 0; index < actualHash.length; index += 1) {
    difference |= actualHash[index] ^ expectedHash[index];
  }

  return difference === 0;
}

export async function proxy(request: NextRequest) {
  const username = process.env.FRONTEND_AUTH_USER ?? "";
  const password = process.env.FRONTEND_AUTH_PASSWORD ?? "";
  if (!username || !password) {
    if (process.env.NODE_ENV === "production") {
      return NextResponse.json(
        { error: { code: "frontend_auth_unconfigured", message: "Frontend authentication is not configured." } },
        { status: 503 },
      );
    }

    return NextResponse.next();
  }

  const authorization = request.headers.get("authorization");
  if (!authorization?.startsWith("Basic ")) return unauthorized();

  let provided = "";
  try {
    const decoded = atob(authorization.slice(6));
    const bytes = Uint8Array.from(decoded, (character) => character.charCodeAt(0));
    provided = new TextDecoder().decode(bytes);
  } catch {
    return unauthorized();
  }
  if (!(await credentialsMatch(provided, `${username}:${password}`))) return unauthorized();

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
