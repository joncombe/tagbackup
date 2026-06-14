import type { FilterMode } from "../types";

interface Props {
  tags: string[];
  selected: Set<string>;
  mode: FilterMode;
  onToggle: (tag: string) => void;
  onModeChange: (mode: FilterMode) => void;
  onClear: () => void;
}

const MODES: { value: FilterMode; label: string }[] = [
  { value: "or", label: "OR" },
  { value: "and", label: "AND" },
];

export function TagFilters({
  tags,
  selected,
  mode,
  onToggle,
  onModeChange,
  onClear,
}: Props) {
  if (tags.length === 0) return null;
  return (
    <div className="tagFilters">
      <div className="tagFilterList">
        {tags.map((t) => (
          <button
            key={t}
            type="button"
            className={`tagFilter${selected.has(t) ? " tagFilterActive" : ""}`}
            aria-pressed={selected.has(t)}
            onClick={() => onToggle(t)}
          >
            {t}
          </button>
        ))}
      </div>
      <div className="tagFilterControls">
        {selected.size > 0 && (
          <button type="button" className="tagClear" onClick={onClear}>
            clear
          </button>
        )}
        <div
          className="modeToggle"
          role="group"
          aria-label="Tag match mode"
        >
          {MODES.map((m) => (
            <button
              key={m.value}
              type="button"
              className={`modeOption${mode === m.value ? " modeOptionActive" : ""}`}
              aria-pressed={mode === m.value}
              onClick={() => onModeChange(m.value)}
            >
              {m.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
