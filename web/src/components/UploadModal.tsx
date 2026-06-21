import { useEffect, useMemo, useState } from "react";
import { uploadObject } from "../api";

interface Props {
  alias: string;
  files: File[];
  existingTags: string[];
  onClose: () => void;
  onComplete: () => void;
}

function isValidTag(tag: string): boolean {
  if (!tag) return false;
  return /^[a-zA-Z0-9]+$/.test(tag);
}

export function UploadModal({
  alias,
  files,
  existingTags,
  onClose,
  onComplete,
}: Props) {
  const [phase, setPhase] = useState<"tagging" | "uploading">("tagging");
  const [selectedTags, setSelectedTags] = useState<Set<string>>(new Set());
  const [draftTag, setDraftTag] = useState("");
  const [tagsToUpload, setTagsToUpload] = useState<string[]>([]);
  const [completed, setCompleted] = useState(0);

  const draftTags = useMemo(() => {
    if (!draftTag.trim()) return [];
    return draftTag
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
  }, [draftTag]);

  const allTags = useMemo(() => {
    const set = new Set(selectedTags);
    for (const t of draftTags) set.add(t);
    return [...set];
  }, [selectedTags, draftTags]);

  const hasInvalidTag = allTags.some((t) => !isValidTag(t));
  const validTags = allTags.filter(isValidTag);
  const canSubmit = validTags.length > 0 && !hasInvalidTag;

  useEffect(() => {
    if (phase !== "uploading") return;
    let cancelled = false;

    async function run() {
      for (let i = 0; i < files.length; i++) {
        if (cancelled) return;
        try {
          await uploadObject(alias, files[i], tagsToUpload);
        } catch {
          // ignore per-file failures and continue
        }
        if (!cancelled) setCompleted(i + 1);
      }
      if (!cancelled) onComplete();
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [alias, files, tagsToUpload, phase, onComplete]);

  function toggleTag(tag: string) {
    setSelectedTags((prev) => {
      const next = new Set(prev);
      if (next.has(tag)) next.delete(tag);
      else next.add(tag);
      return next;
    });
  }

  function startUpload() {
    if (!canSubmit) return;
    setTagsToUpload(validTags);
    setPhase("uploading");
  }

  const count = files.length;
  const progress = count > 0 ? (completed / count) * 100 : 0;

  return (
    <div
      className="modalOverlay"
      role="presentation"
      onClick={phase === "tagging" ? onClose : undefined}
    >
      <div
        className="modalDialog modalDialogWide"
        role="dialog"
        aria-modal="true"
        aria-labelledby="upload-modal-title"
        onClick={(e) => e.stopPropagation()}
      >
        {phase === "tagging" ? (
          <>
            <h2 id="upload-modal-title" className="modalTitle">
              Upload {count} file{count === 1 ? "" : "s"}
            </h2>

            <ul className="uploadFileList">
              {files.map((f) => (
                <li key={`${f.name}-${f.size}-${f.lastModified}`}>{f.name}</li>
              ))}
            </ul>

            <div className="uploadTagSection">
              <p className="uploadTagLabel">Tags</p>
              {existingTags.length > 0 && (
                <div className="tagFilterList">
                  {existingTags.map((t) => (
                    <button
                      key={t}
                      type="button"
                      className={`tagFilter${selectedTags.has(t) ? " tagFilterActive" : ""}`}
                      aria-pressed={selectedTags.has(t)}
                      onClick={() => toggleTag(t)}
                    >
                      {t}
                    </button>
                  ))}
                </div>
              )}
              <input
                type="text"
                className="uploadTagInput"
                placeholder="Add tags (comma-separated)"
                value={draftTag}
                onChange={(e) => setDraftTag(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    startUpload();
                  }
                }}
              />
              {allTags.length > 0 && (
                <div className="tagList uploadSelectedTags">
                  {allTags.map((t) => (
                    <span
                      key={t}
                      className={`tagPill${!isValidTag(t) ? " tagPillInvalid" : ""}`}
                    >
                      {t}
                    </span>
                  ))}
                </div>
              )}
              {hasInvalidTag && (
                <p className="uploadTagError">
                  Tags may only contain letters and numbers (a-z, A-Z, 0-9).
                </p>
              )}
            </div>

            <div className="modalActions">
              <button type="button" className="btnSecondary" onClick={onClose}>
                Cancel
              </button>
              <button
                type="button"
                className="btnPrimary"
                disabled={!canSubmit}
                onClick={startUpload}
              >
                Upload
              </button>
            </div>
          </>
        ) : (
          <>
            <h2 id="upload-modal-title" className="modalTitle">
              Uploading files…
            </h2>
            <p className="modalBody">
              {completed} of {count} uploaded
            </p>
            <div className="progressBar" aria-hidden="true">
              <div className="progressFill" style={{ width: `${progress}%` }} />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
