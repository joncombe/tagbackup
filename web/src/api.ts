import type { FileObject } from "./types";

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url, { headers: { Accept: "application/json" } });
  if (!res.ok) {
    let message = `request failed (${res.status})`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body?.error) message = body.error;
    } catch {
      // response had no JSON body; keep the default message
    }
    throw new Error(message);
  }
  return (await res.json()) as T;
}

export function fetchBuckets(): Promise<string[]> {
  return getJSON<string[]>("/api/buckets");
}

export function fetchObjects(alias: string): Promise<FileObject[]> {
  return getJSON<FileObject[]>(
    `/api/buckets/${encodeURIComponent(alias)}/objects`,
  );
}
