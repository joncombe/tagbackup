export interface FileObject {
  key: string;
  filename: string;
  tags: string[];
  size: number;
  /** Epoch milliseconds. */
  timestamp: number;
}

export type SortField = "filename" | "size" | "timestamp";
export type SortDirection = "asc" | "desc";

export interface BucketConfig {
  alias: string;
  bucket: string;
  endpoint: string;
  region: string;
  prefix?: string;
  force_path_style: boolean;
  insecure_skip_verify: boolean;
  credential_type: string;
  credential_source: "env" | "static" | "profile" | "iam";
  access_key_id?: string;
  secret_access_key?: string;
  credentials_profile?: string;
}
