interface Props {
  buckets: string[];
  selected: string | null;
  onSelect: (bucket: string) => void;
}

export function BucketTabs({ buckets, selected, onSelect }: Props) {
  return (
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
  );
}
