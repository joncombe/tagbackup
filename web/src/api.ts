import type { BucketConfig, FileObject } from "./types";

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

async function parseError(res: Response): Promise<string> {
  let message = `request failed (${res.status})`;
  try {
    const body = (await res.json()) as { error?: string };
    if (body?.error) message = body.error;
  } catch {
    // response had no JSON body; keep the default message
  }
  return message;
}

export function fetchBuckets(): Promise<string[]> {
  return getJSON<string[]>("/api/buckets");
}

export async function fetchVersion(): Promise<string> {
  const data = await getJSON<{ version: string }>("/api/version");
  return data.version;
}

export function fetchBucketConfig(alias: string): Promise<BucketConfig> {
  return getJSON<BucketConfig>(`/api/buckets/${encodeURIComponent(alias)}`);
}

export function fetchObjects(alias: string): Promise<FileObject[]> {
  return getJSON<FileObject[]>(
    `/api/buckets/${encodeURIComponent(alias)}/objects`,
  );
}

export async function deleteObject(alias: string, key: string): Promise<void> {
  const res = await fetch(`/api/buckets/${encodeURIComponent(alias)}/objects`, {
    method: "DELETE",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ key }),
  });
  if (!res.ok) {
    throw new Error(await parseError(res));
  }
}

export async function uploadObject(
  alias: string,
  file: File,
  tags: string[],
): Promise<FileObject> {
  const form = new FormData();
  form.append("file", file);
  form.append("tags", JSON.stringify(tags));
  const res = await fetch(`/api/buckets/${encodeURIComponent(alias)}/objects`, {
    method: "POST",
    body: form,
  });
  if (!res.ok) {
    throw new Error(await parseError(res));
  }
  return (await res.json()) as FileObject;
}
