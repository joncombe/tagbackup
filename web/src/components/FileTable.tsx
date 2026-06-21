import { useEffect, useRef } from "react";
import { ChevronDown, ChevronUp, Download, Tag } from "lucide-react";
import { objectDownloadURL } from "../api";
import type { FileObject, SortDirection, SortField } from "../types";
import { formatTimestamp, humanBytes, relativeTime } from "../format";

interface Props {
  rows: FileObject[];
  bucketAlias: string;
  sortField: SortField;
  sortDirection: SortDirection;
  onSort: (field: SortField) => void;
  now: number;
  checkedKeys: Set<string>;
  onCheck: (key: string, checked: boolean) => void;
  onCheckAll: (checked: boolean) => void;
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

function SortIndicator({
  active,
  direction,
}: {
  active: boolean;
  direction: SortDirection;
}) {
  if (!active) return null;
  const Icon = direction === "asc" ? ChevronUp : ChevronDown;
  return <Icon className="sortIcon" size={14} aria-hidden="true" />;
}

export function FileTable({
  rows,
  bucketAlias,
  sortField,
  sortDirection,
  onSort,
  now,
  checkedKeys,
  onCheck,
  onCheckAll,
}: Props) {
  const selectAllRef = useRef<HTMLInputElement>(null);
  const checkedCount = rows.filter((r) => checkedKeys.has(r.key)).length;
  const allChecked = rows.length > 0 && checkedCount === rows.length;
  const someChecked = checkedCount > 0 && !allChecked;

  useEffect(() => {
    if (selectAllRef.current) {
      selectAllRef.current.indeterminate = someChecked;
    }
  }, [someChecked]);

  function toggleSelectAll() {
    onCheckAll(!allChecked);
  }

  return (
    <div className="fileTableScroll">
      <table className="fileTable">
        <thead>
          <tr>
            <th className="colCheck">
              <input
                ref={selectAllRef}
                type="checkbox"
                className="rowCheck"
                checked={allChecked}
                aria-label="Select all files on this page"
                onChange={toggleSelectAll}
              />
            </th>
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
                    <SortIndicator active={active} direction={sortDirection} />
                  </button>
                </th>
              );
            })}
            <th className="colTags">Tags</th>
            <th className="colDownload" aria-label="Download" />
          </tr>
        </thead>
        <tbody>
          {rows.map((o) => (
            <tr key={o.key}>
              <td className="colCheck">
                <input
                  type="checkbox"
                  className="rowCheck"
                  checked={checkedKeys.has(o.key)}
                  aria-label={`Select ${o.filename}`}
                  onChange={(e) => onCheck(o.key, e.target.checked)}
                />
              </td>
              <td className="cellName">{o.filename}</td>
              <td className="colSize">{humanBytes(o.size)}</td>
              <td className="colTime">
                <span className="absTime">{formatTimestamp(o.timestamp)}</span>
                <span className="relTime">
                  {relativeTime(o.timestamp, now)}
                </span>
              </td>
              <td className="colTags">
                <span className="tagList">
                  {o.tags.map((t) => (
                    <span key={t} className="tagPill">
                      <Tag size={12} aria-hidden="true" />
                      {t}
                    </span>
                  ))}
                </span>
              </td>
              <td className="colDownload">
                <a
                  href={objectDownloadURL(bucketAlias, o.key)}
                  download={o.filename}
                  className="downloadBtn"
                  aria-label={`Download ${o.filename}`}
                >
                  <Download size={16} aria-hidden="true" />
                </a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
