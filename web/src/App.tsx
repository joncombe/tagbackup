import { useEffect, useMemo, useState } from "react";
import { fetchBuckets, fetchObjects } from "./api";
import type {
  FileObject,
  FilterMode,
  SortDirection,
  SortField,
} from "./types";
import { BucketTabs } from "./components/BucketTabs";
import { TagFilters } from "./components/TagFilters";
import { FileTable } from "./components/FileTable";
import { Pagination } from "./components/Pagination";

const PAGE_SIZE_OPTIONS = [25, 50, 100, 250];

function defaultDirection(field: SortField): SortDirection {
  // Newest-first and largest-first read more naturally as the initial sort.
  return field === "filename" ? "asc" : "desc";
}

export function App() {
  const [buckets, setBuckets] = useState<string[] | null>(null);
  const [bucketsError, setBucketsError] = useState<string | null>(null);
  const [selected, setSelected] = useState<string | null>(null);

  const [objects, setObjects] = useState<FileObject[]>([]);
  const [loadingObjects, setLoadingObjects] = useState(false);
  const [objectsError, setObjectsError] = useState<string | null>(null);

  const [selectedTags, setSelectedTags] = useState<Set<string>>(new Set());
  const [filterMode, setFilterMode] = useState<FilterMode>("or");
  const [sortField, setSortField] = useState<SortField>("timestamp");
  const [sortDirection, setSortDirection] = useState<SortDirection>("desc");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(PAGE_SIZE_OPTIONS[0]);

  // A single "now" snapshot keeps relative times stable across a render and
  // refreshes them every minute.
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 60_000);
    return () => clearInterval(id);
  }, []);

  useEffect(() => {
    fetchBuckets()
      .then((list) => {
        setBuckets(list);
        if (list.length > 0) setSelected(list[0]);
      })
      .catch((e: unknown) =>
        setBucketsError(e instanceof Error ? e.message : String(e)),
      );
  }, []);

  useEffect(() => {
    if (!selected) return;
    let cancelled = false;
    setLoadingObjects(true);
    setObjectsError(null);
    setSelectedTags(new Set());
    setPage(1);
    fetchObjects(selected)
      .then((list) => {
        if (!cancelled) setObjects(list);
      })
      .catch((e: unknown) => {
        if (!cancelled)
          setObjectsError(e instanceof Error ? e.message : String(e));
      })
      .finally(() => {
        if (!cancelled) setLoadingObjects(false);
      });
    return () => {
      cancelled = true;
    };
  }, [selected]);

  const tags = useMemo(() => {
    const set = new Set<string>();
    for (const o of objects) for (const t of o.tags) set.add(t);
    return [...set].sort();
  }, [objects]);

  const filtered = useMemo(() => {
    if (selectedTags.size === 0) return objects;
    return objects.filter((o) => {
      const have = new Set(o.tags);
      if (filterMode === "or") {
        for (const t of selectedTags) if (have.has(t)) return true;
        return false;
      }
      for (const t of selectedTags) if (!have.has(t)) return false;
      return true;
    });
  }, [objects, selectedTags, filterMode]);

  const sorted = useMemo(() => {
    const copy = [...filtered];
    const dir = sortDirection === "asc" ? 1 : -1;
    copy.sort((a, b) => {
      let cmp = 0;
      if (sortField === "filename") {
        cmp = a.filename.localeCompare(b.filename);
      } else if (sortField === "size") {
        cmp = a.size - b.size;
      } else {
        cmp = a.timestamp - b.timestamp;
      }
      if (cmp === 0) cmp = a.key.localeCompare(b.key);
      return cmp * dir;
    });
    return copy;
  }, [filtered, sortField, sortDirection]);

  const pageCount = Math.max(Math.ceil(sorted.length / pageSize), 1);
  const currentPage = Math.min(page, pageCount);
  const pageRows = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return sorted.slice(start, start + pageSize);
  }, [sorted, currentPage, pageSize]);

  function toggleTag(tag: string) {
    setSelectedTags((prev) => {
      const next = new Set(prev);
      if (next.has(tag)) next.delete(tag);
      else next.add(tag);
      return next;
    });
    setPage(1);
  }

  function changeFilterMode(mode: FilterMode) {
    setFilterMode(mode);
    setPage(1);
  }

  function onSort(field: SortField) {
    if (field === sortField) {
      setSortDirection((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortField(field);
      setSortDirection(defaultDirection(field));
    }
    setPage(1);
  }

  return (
    <div className="app">
      <header className="topbar">
        <span className="wordmark">tagbackup</span>
        <span className="readonlyBadge">read-only</span>
      </header>

      {buckets !== null && buckets.length > 0 && (
        <BucketTabs
          buckets={buckets}
          selected={selected}
          onSelect={setSelected}
        />
      )}

      <main className="content">
        {bucketsError && (
          <div className="empty error">Failed to load buckets: {bucketsError}</div>
        )}

        {!bucketsError && buckets !== null && buckets.length === 0 && (
          <div className="empty">
            <h2>No buckets configured</h2>
            <p>
              Add one with <code>tagbackup bucket add</code>, then refresh this
              page.
            </p>
          </div>
        )}

        {selected && (
          <>
            <TagFilters
              tags={tags}
              selected={selectedTags}
              mode={filterMode}
              onToggle={toggleTag}
              onModeChange={changeFilterMode}
              onClear={() => {
                setSelectedTags(new Set());
                setPage(1);
              }}
            />

            {loadingObjects && <div className="empty">Loading files…</div>}

            {!loadingObjects && objectsError && (
              <div className="empty error">
                Failed to load files: {objectsError}
              </div>
            )}

            {!loadingObjects && !objectsError && sorted.length === 0 && (
              <div className="empty">
                {objects.length === 0
                  ? "No files in this bucket."
                  : "No files match the selected tags."}
              </div>
            )}

            {!loadingObjects && !objectsError && sorted.length > 0 && (
              <FileTable
                rows={pageRows}
                sortField={sortField}
                sortDirection={sortDirection}
                onSort={onSort}
                now={now}
              />
            )}
          </>
        )}
      </main>

      {selected && !loadingObjects && !objectsError && (
        <Pagination
          page={currentPage}
          pageCount={pageCount}
          pageSize={pageSize}
          total={sorted.length}
          pageSizeOptions={PAGE_SIZE_OPTIONS}
          onPageChange={setPage}
          onPageSizeChange={(s) => {
            setPageSize(s);
            setPage(1);
          }}
        />
      )}
    </div>
  );
}
