interface Props {
  buckets: string[];
  selected: string | null;
  onSelect: (bucket: string) => void;
  onInfoClick: () => void;
}

export function BucketTabs({
  buckets,
  selected,
  onSelect,
  onInfoClick,
}: Props) {
  return (
    <div className="tabsRow">
      <nav className="tabs" aria-label="Buckets">
        {buckets.map((b) => (
          <button
            key={b}
            type="button"
            className={`tab${b === selected ? " tabActive" : ""}`}
            aria-current={b === selected ? "page" : undefined}
            onClick={() => onSelect(b)}
          >
            {b}
          </button>
        ))}
      </nav>
      {selected && (
        <button
          type="button"
          className="tabInfoBtn"
          aria-label="View bucket configuration"
          onClick={onInfoClick}
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            aria-hidden="true"
          >
            <circle
              cx="8"
              cy="8"
              r="7"
              stroke="currentColor"
              strokeWidth="1.25"
            />
            <path
              d="M8 7.25V11"
              stroke="currentColor"
              strokeWidth="1.25"
              strokeLinecap="round"
            />
            <circle cx="8" cy="5.25" r="0.75" fill="currentColor" />
          </svg>
        </button>
      )}
    </div>
  );
}
