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
export type FilterMode = "or" | "and";
