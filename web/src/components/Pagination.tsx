import { ChevronLeft, ChevronRight } from "lucide-react";
import { humanBytes } from "../format";

interface Props {
  page: number;
  pageCount: number;
  total: number;
  totalBytes: number;
  onPageChange: (page: number) => void;
  appVersion: string | null;
  showPageData: boolean;
}

export function Pagination({
  page,
  pageCount,
  total,
  totalBytes,
  onPageChange,
  appVersion,
  showPageData,
}: Props) {
  const safePageCount = Math.max(pageCount, 1);
  return (
    <footer className="pagination">
      <div className="paginationScroll">
        <div className="pageInfo">
          {showPageData &&
            (total === 0
              ? "No files"
              : `${total} file${total === 1 ? "" : "s"} · ${humanBytes(totalBytes)}`)}
        </div>
        <div className="footerMeta">
          {appVersion && <span className="footerVersion">v{appVersion}</span>}
          <a
            className="footerLink"
            href="https://tagbackup.com/"
            target="_blank"
            rel="noopener noreferrer"
          >
            tagbackup.com
          </a>
        </div>
        <div className="pageControls">
          {showPageData && (
            <>
              <button
                type="button"
                className="pageBtn"
                disabled={page <= 1}
                onClick={() => onPageChange(page - 1)}
                aria-label="Previous page"
              >
                <ChevronLeft size={16} aria-hidden="true" />
              </button>
              <span className="pageStatus">
                {Math.min(page, safePageCount)} of {safePageCount}
              </span>
              <button
                type="button"
                className="pageBtn"
                disabled={page >= safePageCount}
                onClick={() => onPageChange(page + 1)}
                aria-label="Next page"
              >
                <ChevronRight size={16} aria-hidden="true" />
              </button>
            </>
          )}
        </div>
      </div>
    </footer>
  );
}
