import { Check } from "lucide-react";

interface Props {
  onClose: () => void;
}

export function DeleteHintModal({ onClose }: Props) {
  return (
    <div className="modalOverlay" role="presentation" onClick={onClose}>
      <div
        className="modalDialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="delete-hint-modal-title"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="delete-hint-modal-title" className="modalTitle">
          No files selected
        </h2>
        <p className="modalBody">
          Select one or more files using the checkboxes in the file list, then
          click Delete again.
        </p>
        <div className="modalActions">
          <button type="button" className="btnSecondary" onClick={onClose}>
            <Check size={14} aria-hidden="true" />
            OK
          </button>
        </div>
      </div>
    </div>
  );
}
