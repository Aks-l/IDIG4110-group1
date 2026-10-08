import type { NextRequest } from 'next/server';

/**
 * Shared plumbing for the placeholder routes in src/app/api.
 *
 * The placeholder routes implement the frontend contract from
 * docs/frontend/api.md with fixture data until the real backend exposes the
 * same endpoints. Swapping in the real backend is configuration-only:
 *
 * 1. Point the browser straight at it: set NEXT_PUBLIC_API_URL to the real
 *    base URL (for example http://localhost:8080/api). The placeholder
 *    routes are then simply bypassed.
 *
 * 2. Keep the browser on the placeholder routes: set PLACEHOLDER_API_TARGET
 *    (server-side only) to the real base URL, and every placeholder route
 *    forwards its request there instead of serving fixtures.
 */

const API_PREFIX = '/api';

export function placeholderTarget(): string {
  return (process.env.PLACEHOLDER_API_TARGET ?? '').trim().replace(/\/+$/, '');
}

export async function handlePlaceholderRequest(
  request: NextRequest,
  serveFixtures: () => Response | Promise<Response>,
): Promise<Response> {
  const target = placeholderTarget();
  if (!target) return serveFixtures();

  const incoming = request.nextUrl;
  const path = incoming.pathname.slice(API_PREFIX.length) || '/';
  const upstream = new URL(target + path + incoming.search);

  // Never forward to ourselves: a PLACEHOLDER_API_TARGET pointing at this
  // app's own origin would forward the request back to this handler forever.
  if (upstream.href === incoming.href) return serveFixtures();

  try {
    const headers = new Headers();
    for (const header of ['content-type', 'authorization']) {
      const value = request.headers.get(header);
      if (value) headers.set(header, value);
    }

    const hasBody = request.method !== 'GET' && request.method !== 'HEAD';
    const response = await fetch(upstream, {
      method: request.method,
      headers,
      body: hasBody ? await request.text() : undefined,
      cache: 'no-store',
    });

    return new Response(response.body, {
      status: response.status,
      headers: {
        'content-type': response.headers.get('content-type') ?? 'application/json',
        'cache-control': 'no-store',
      },
    });
  } catch {
    return jsonError(502, `Placeholder API target is unreachable: ${target}`);
  }
}

export function jsonOk(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: { 'content-type': 'application/json', 'cache-control': 'no-store' },
  });
}

export function jsonError(status: number, message: string): Response {
  // Same shape the Go services return, so clients can always read
  // { code, message } from an error response.
  return new Response(JSON.stringify({ code: status, message }), {
    status,
    headers: { 'content-type': 'application/json', 'cache-control': 'no-store' },
  });
}

/**
 * Parses a JSON object body, returning {} for malformed input so callers
 * can respond with a 400 instead of crashing.
 */
export async function readJsonObject(
  request: NextRequest,
): Promise<Record<string, unknown>> {
  try {
    const body: unknown = await request.json();
    if (body !== null && typeof body === 'object' && !Array.isArray(body)) {
      return body as Record<string, unknown>;
    }
  } catch {
    // fall through: malformed or non-object bodies become {}
  }
  return {};
}
