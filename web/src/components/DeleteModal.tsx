import { useEffect, useState } from "react";
import { deleteObject } from "../api";

interface Props {
  alias: string;
  keys: string[];
  onClose: () => void;
  onComplete: () => void;
}

export function DeleteModal({ alias, keys, onClose, onComplete }: Props) {
  const [phase, setPhase] = useState<"confirm" | "deleting">("confirm");
  const [completed, setCompleted] = useState(0);

  useEffect(() => {
    if (phase !== "deleting") return;
    let cancelled = false;

    async function run() {
      for (let i = 0; i < keys.length; i++) {
        if (cancelled) return;
        try {
          await deleteObject(alias, keys[i]);
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
  }, [alias, keys, phase, onComplete]);

  const count = keys.length;
  const progress = count > 0 ? (completed / count) * 100 : 0;

  return (
    <div
      className="modalOverlay"
      role="presentation"
      onClick={phase === "confirm" ? onClose : undefined}
    >
      <div
        className="modalDialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="delete-modal-title"
        onClick={(e) => e.stopPropagation()}
      >
        {phase === "confirm" ? (
          <>
            <h2 id="delete-modal-title" className="modalTitle">
              Delete {count} file{count === 1 ? "" : "s"}?
            </h2>
            <p className="modalBody">
              This will permanently remove the selected file
              {count === 1 ? "" : "s"} from the bucket. This action cannot be
              undone.
            </p>
            <div className="modalActions">
              <button type="button" className="btnSecondary" onClick={onClose}>
                Cancel
              </button>
              <button
                type="button"
                className="btnDanger"
                onClick={() => setPhase("deleting")}
              >
                Delete
              </button>
            </div>
          </>
        ) : (
          <>
            <h2 id="delete-modal-title" className="modalTitle">
              Deleting files…
            </h2>
            <p className="modalBody">
              {completed} of {count} deleted
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
