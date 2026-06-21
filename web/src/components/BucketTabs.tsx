import { HardDrive, Info } from "lucide-react";

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
            <HardDrive size={14} aria-hidden="true" />
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
          <Info size={16} aria-hidden="true" />
        </button>
      )}
    </div>
  );
}
