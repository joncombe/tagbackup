import type { FileObject, SortDirection, SortField } from "../types";
import { formatTimestamp, humanBytes, relativeTime } from "../format";

interface Props {
  rows: FileObject[];
  sortField: SortField;
  sortDirection: SortDirection;
  onSort: (field: SortField) => void;
  now: number;
}

interface ColumnDef {
  field: SortField;
  label: string;
  className?: string;
}

const COLUMNS: ColumnDef[] = [
  { field: "filename", label: "Filename" },
  { field: "size", label: "Size", className: "colSize" },
  { field: "timestamp", label: "Timestamp", className: "colTime" },
];

function sortIndicator(active: boolean, dir: SortDirection): string {
  if (!active) return "";
  return dir === "asc" ? " \u2191" : " \u2193";
}

export function FileTable({ rows, sortField, sortDirection, onSort, now }: Props) {
  return (
    <table className="fileTable">
      <thead>
        <tr>
          {COLUMNS.map((c) => {
            const active = sortField === c.field;
            return (
              <th key={c.field} className={c.className}>
                <button
                  type="button"
                  className={`sortHeader${active ? " sortHeaderActive" : ""}`}
                  aria-sort={
                    active
                      ? sortDirection === "asc"
                        ? "ascending"
                        : "descending"
                      : "none"
                  }
                  onClick={() => onSort(c.field)}
                >
                  {c.label}
                  {sortIndicator(active, sortDirection)}
                </button>
              </th>
            );
          })}
          <th className="colTags">Tags</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((o) => (
          <tr key={o.key}>
            <td className="cellName">{o.filename}</td>
            <td className="colSize">{humanBytes(o.size)}</td>
            <td className="colTime">
              <span className="absTime">{formatTimestamp(o.timestamp)}</span>
              <span className="relTime">{relativeTime(o.timestamp, now)}</span>
            </td>
            <td className="colTags">
              <span className="tagList">
                {o.tags.map((t) => (
                  <span key={t} className="tagPill">
                    {t}
                  </span>
                ))}
              </span>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
