import { useEffect, useState } from "react";
import { fetchBucketConfig } from "../api";
import type { BucketConfig } from "../types";

interface Props {
  alias: string;
  onClose: () => void;
}

const FIELD_LABELS: Record<string, string> = {
  alias: "Alias",
  bucket: "Bucket",
  endpoint: "Endpoint",
  region: "Region",
  prefix: "Prefix",
  force_path_style: "Force path style",
  insecure_skip_verify: "Insecure skip verify",
  credential_type: "Credential type",
  credential_source: "Credential source",
  access_key_id: "Access key ID",
  secret_access_key: "Secret access key",
  credentials_profile: "Credentials profile",
};

const FIELD_ORDER: (keyof BucketConfig)[] = [
  "alias",
  "bucket",
  "endpoint",
  "region",
  "prefix",
  "force_path_style",
  "insecure_skip_verify",
  "credential_type",
  "credential_source",
  "credentials_profile",
  "access_key_id",
  "secret_access_key",
];

function formatValue(value: unknown): string {
  if (typeof value === "boolean") return value ? "true" : "false";
  if (value === undefined || value === null || value === "") return "(not set)";
  return String(value);
}

export function BucketInfoModal({ alias, onClose }: Props) {
  const [config, setConfig] = useState<BucketConfig | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    setConfig(null);

    void fetchBucketConfig(alias)
      .then((data) => {
        if (!cancelled) setConfig(data);
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(
            err instanceof Error ? err.message : "Failed to load config",
          );
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [alias]);

  return (
    <div className="modalOverlay" role="presentation" onClick={onClose}>
      <div
        className="modalDialog modalDialogWide"
        role="dialog"
        aria-modal="true"
        aria-labelledby="bucket-info-modal-title"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="bucket-info-modal-title" className="modalTitle">
          Bucket: {alias}
        </h2>

        {loading && <p className="modalBody">Loading configuration…</p>}

        {error && (
          <p className="modalBody error">
            Failed to load configuration: {error}
          </p>
        )}

        {!loading && !error && config && (
          <>
            {config.credential_source === "env" && (
              <p className="configEnvNote">
                Credentials are provided via environment variables (values not
                shown).
              </p>
            )}
            <dl className="configList">
              {FIELD_ORDER.map((key) => {
                const value = config[key];
                if (
                  value === undefined ||
                  value === null ||
                  (typeof value === "string" && value === "")
                ) {
                  if (
                    key === "prefix" ||
                    key === "credentials_profile" ||
                    key === "access_key_id" ||
                    key === "secret_access_key"
                  ) {
                    return null;
                  }
                }
                return (
                  <div key={key} className="configRow">
                    <dt className="configLabel">{FIELD_LABELS[key]}</dt>
                    <dd className="configValue">{formatValue(value)}</dd>
                  </div>
                );
              })}
            </dl>
          </>
        )}

        <div className="modalActions">
          <button type="button" className="btnSecondary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
