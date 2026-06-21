interface Props {
  tags: string[];
  selected: string | null;
  onSelect: (tag: string) => void;
}

export function TagFilters({ tags, selected, onSelect }: Props) {
  if (tags.length === 0) return null;
  return (
    <div className="tagFilters">
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
    </div>
  );
}
