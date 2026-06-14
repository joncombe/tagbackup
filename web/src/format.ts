const KIB = 1024;
const MIB = KIB * 1024;
const GIB = MIB * 1024;
const TIB = GIB * 1024;

/** humanBytes mirrors the CLI's binary (KiB/MiB/...) formatting. */
export function humanBytes(n: number): string {
  if (n < 0) return "-";
  if (n < KIB) return `${n} B`;
  if (n < MIB) return `${(n / KIB).toFixed(1)} KiB`;
  if (n < GIB) return `${(n / MIB).toFixed(1)} MiB`;
  if (n < TIB) return `${(n / GIB).toFixed(1)} GiB`;
  return `${(n / TIB).toFixed(1)} TiB`;
}

/** formatTimestamp renders an absolute local date-time. */
export function formatTimestamp(ms: number): string {
  const d = new Date(ms);
  const pad = (x: number) => String(x).padStart(2, "0");
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}`
  );
}

const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 1000 * 60 * 60 * 24 * 365],
  ["month", 1000 * 60 * 60 * 24 * 30],
  ["day", 1000 * 60 * 60 * 24],
  ["hour", 1000 * 60 * 60],
  ["minute", 1000 * 60],
  ["second", 1000],
];

const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });

/** relativeTime renders e.g. "6 hours ago" relative to now. */
export function relativeTime(ms: number, now: number = Date.now()): string {
  const diff = ms - now;
  const abs = Math.abs(diff);
  for (const [unit, unitMs] of UNITS) {
    if (abs >= unitMs || unit === "second") {
      const value = Math.round(diff / unitMs);
      return rtf.format(value, unit);
    }
  }
  return rtf.format(0, "second");
}
