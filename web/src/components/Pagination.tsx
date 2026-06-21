import { humanBytes } from "../format";

interface Props {
  page: number;
  pageCount: number;
  total: number;
  totalBytes: number;
  onPageChange: (page: number) => void;
}

export function Pagination({
  page,
  pageCount,
  total,
  totalBytes,
  onPageChange,
}: Props) {
  const safePageCount = Math.max(pageCount, 1);
  return (
    <footer className="pagination">
      <div className="pageInfo">
        {total === 0
          ? "No files"
          : `${total} file${total === 1 ? "" : "s"} · ${humanBytes(totalBytes)}`}
      </div>
      <div className="pageControls">
        <button
          type="button"
          className="pageBtn"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
          aria-label="Previous page"
        >
          &larr;
        </button>
        <span className="pageStatus">
          Page {Math.min(page, safePageCount)} of {safePageCount}
        </span>
        <button
          type="button"
          className="pageBtn"
          disabled={page >= safePageCount}
          onClick={() => onPageChange(page + 1)}
          aria-label="Next page"
        >
          &rarr;
        </button>
      </div>
    </footer>
  );
}
