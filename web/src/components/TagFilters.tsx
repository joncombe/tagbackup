interface Props {
  tags: string[];
  selected: string | null;
  onSelect: (tag: string) => void;
  checkedCount: number;
  onUploadClick: () => void;
  onDeleteClick: () => void;
}

export function TagFilters({
  tags,
  selected,
  onSelect,
  checkedCount,
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
              {t}
            </button>
          ))}
        </div>
      )}
      <div className="filterActions">
        <button type="button" className="btnUpload" onClick={onUploadClick}>
          Upload
        </button>
        <button
          type="button"
          className="btnDelete"
          disabled={checkedCount === 0}
          onClick={onDeleteClick}
        >
          Delete
        </button>
      </div>
    </div>
  );
}
