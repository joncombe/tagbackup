import { useCallback, useEffect, useMemo, useState } from "react";
import { fetchBuckets, fetchObjects, fetchVersion } from "./api";
import type { FileObject, SortDirection, SortField } from "./types";
import { BucketTabs } from "./components/BucketTabs";
import { TagFilters } from "./components/TagFilters";
import { FileTable } from "./components/FileTable";
import { Pagination } from "./components/Pagination";
import { DeleteModal } from "./components/DeleteModal";
import { DeleteHintModal } from "./components/DeleteHintModal";
import { BucketInfoModal } from "./components/BucketInfoModal";
import { UploadDropZone } from "./components/UploadDropZone";
import { UploadModal } from "./components/UploadModal";

const PAGE_SIZE = 50;

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

  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [sortField, setSortField] = useState<SortField>("timestamp");

  useEffect(() => {
    document.title = selected ? `tagbackup | ${selected}` : "tagbackup";
  }, [selected]);
  const [sortDirection, setSortDirection] = useState<SortDirection>("desc");
  const [page, setPage] = useState(1);

  const [checkedKeys, setCheckedKeys] = useState<Set<string>>(new Set());
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showDeleteHintModal, setShowDeleteHintModal] = useState(false);
  const [showDropZone, setShowDropZone] = useState(false);
  const [pendingFiles, setPendingFiles] = useState<File[]>([]);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [showInfoModal, setShowInfoModal] = useState(false);
  const [appVersion, setAppVersion] = useState<string | null>(null);

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
    void fetchVersion()
      .then(setAppVersion)
      .catch(() => {
        // omit version if unavailable
      });
  }, []);

  const reloadObjects = useCallback(async (alias: string) => {
    setLoadingObjects(true);
    setObjectsError(null);
    try {
      const list = await fetchObjects(alias);
      setObjects(list);
    } catch (e: unknown) {
      setObjectsError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoadingObjects(false);
    }
  }, []);

  useEffect(() => {
    if (!selected) return;
    let cancelled = false;
    setLoadingObjects(true);
    setObjectsError(null);
    setSelectedTag(null);
    setPage(1);
    setCheckedKeys(new Set());
    setShowDropZone(false);
    setPendingFiles([]);
    setShowUploadModal(false);
    setShowDeleteModal(false);
    setShowDeleteHintModal(false);
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

  useEffect(() => {
    setCheckedKeys(new Set());
  }, [selectedTag, page]);

  const tags = useMemo(() => {
    const set = new Set<string>();
    for (const o of objects) for (const t of o.tags) set.add(t);
    return [...set].sort();
  }, [objects]);

  useEffect(() => {
    if (selectedTag !== null && !tags.includes(selectedTag)) {
      setSelectedTag(null);
      setPage(1);
    }
  }, [tags, selectedTag]);

  const filtered = useMemo(() => {
    if (selectedTag === null) return objects;
    return objects.filter((o) => o.tags.includes(selectedTag));
  }, [objects, selectedTag]);

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

  const totalBytes = useMemo(
    () => filtered.reduce((sum, o) => sum + o.size, 0),
    [filtered],
  );

  const pageCount = Math.max(Math.ceil(sorted.length / PAGE_SIZE), 1);
  const currentPage = Math.min(page, pageCount);
  const pageRows = useMemo(() => {
    const start = (currentPage - 1) * PAGE_SIZE;
    return sorted.slice(start, start + PAGE_SIZE);
  }, [sorted, currentPage]);

  const keysToDelete = [...checkedKeys];

  function selectTag(tag: string) {
    setSelectedTag((prev) => (prev === tag ? null : tag));
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

  function changePage(next: number) {
    setPage(next);
    window.scrollTo(0, 0);
  }

  function handleCheck(key: string, checked: boolean) {
    setCheckedKeys((prev) => {
      const next = new Set(prev);
      if (checked) next.add(key);
      else next.delete(key);
      return next;
    });
  }

  function handleCheckAll(checked: boolean) {
    if (!checked) {
      setCheckedKeys(new Set());
      return;
    }
    setCheckedKeys(new Set(pageRows.map((r) => r.key)));
  }

  function handleUploadClick() {
    if (showDropZone && pendingFiles.length === 0) {
      setShowDropZone(false);
    } else {
      setShowDropZone(true);
    }
  }

  function handleFilesSelected(files: File[]) {
    setPendingFiles(files);
    setShowUploadModal(true);
  }

  function handleDeleteClick() {
    if (checkedKeys.size === 0) {
      setShowDeleteHintModal(true);
      return;
    }
    setShowDeleteModal(true);
  }

  async function handleDeleteComplete() {
    setShowDeleteModal(false);
    setCheckedKeys(new Set());
    if (selected) await reloadObjects(selected);
  }

  async function handleUploadComplete() {
    setShowUploadModal(false);
    setShowDropZone(false);
    setPendingFiles([]);
    if (selected) await reloadObjects(selected);
  }

  return (
    <div className="app">
      <header className="topbar">
        <span className="wordmark">tagbackup</span>
      </header>

      {buckets !== null && buckets.length > 0 && (
        <BucketTabs
          buckets={buckets}
          selected={selected}
          onSelect={setSelected}
          onInfoClick={() => setShowInfoModal(true)}
        />
      )}

      <main className="content">
        {bucketsError && (
          <div className="empty error">
            Failed to load buckets: {bucketsError}
          </div>
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
              selected={selectedTag}
              onSelect={selectTag}
              uploadActive={showDropZone}
              onUploadClick={handleUploadClick}
              onDeleteClick={handleDeleteClick}
            />

            {showDropZone && (
              <UploadDropZone onFilesSelected={handleFilesSelected} />
            )}

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
                  : "No files match the selected tag."}
              </div>
            )}

            {!loadingObjects && !objectsError && sorted.length > 0 && (
              <FileTable
                rows={pageRows}
                bucketAlias={selected}
                sortField={sortField}
                sortDirection={sortDirection}
                onSort={onSort}
                now={now}
                checkedKeys={checkedKeys}
                onCheck={handleCheck}
                onCheckAll={handleCheckAll}
              />
            )}
          </>
        )}
      </main>

      <Pagination
        page={currentPage}
        pageCount={pageCount}
        total={sorted.length}
        totalBytes={totalBytes}
        onPageChange={changePage}
        appVersion={appVersion}
        showPageData={Boolean(selected && !loadingObjects && !objectsError)}
      />

      {showInfoModal && selected && (
        <BucketInfoModal
          alias={selected}
          onClose={() => setShowInfoModal(false)}
        />
      )}

      {showDeleteHintModal && (
        <DeleteHintModal onClose={() => setShowDeleteHintModal(false)} />
      )}

      {showDeleteModal && selected && (
        <DeleteModal
          alias={selected}
          keys={keysToDelete}
          onClose={() => setShowDeleteModal(false)}
          onComplete={() => void handleDeleteComplete()}
        />
      )}

      {showUploadModal && selected && pendingFiles.length > 0 && (
        <UploadModal
          alias={selected}
          files={pendingFiles}
          existingTags={tags}
          onClose={() => {
            setShowUploadModal(false);
            setPendingFiles([]);
          }}
          onComplete={() => void handleUploadComplete()}
        />
      )}
    </div>
  );
}
