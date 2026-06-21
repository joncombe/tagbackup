import { Tag, Trash2, Upload } from "lucide-react";

interface Props {
  tags: string[];
  selected: string | null;
  onSelect: (tag: string) => void;
  uploadActive: boolean;
  onUploadClick: () => void;
  onDeleteClick: () => void;
}

export function TagFilters({
  tags,
  selected,
  onSelect,
  uploadActive,
  onUploadClick,
  onDeleteClick,
}: Props) {
  return (
    <div className="tagFilters">
      {tags.length > 0 && (
        <div className="tagFilterList">
          {tags.map((t) => (
            <button
              key={t}
              type="button"
              className={`tagFilter${selected === t ? " tagFilterActive" : ""}`}
              aria-pressed={selected === t}
              onClick={() => onSelect(t)}
            >
              <Tag size={12} aria-hidden="true" />
              {t}
            </button>
          ))}
        </div>
      )}
      <div className="filterActions">
        <button
          type="button"
          className={`btnUpload${uploadActive ? " btnUploadActive" : ""}`}
          aria-pressed={uploadActive}
          onClick={onUploadClick}
        >
          <Upload size={14} aria-hidden="true" />
          Upload
        </button>
        <button type="button" className="btnDelete" onClick={onDeleteClick}>
          <Trash2 size={14} aria-hidden="true" />
          Delete
        </button>
      </div>
    </div>
  );
}
