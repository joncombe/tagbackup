import { useRef, useState } from "react";

interface Props {
  onFilesSelected: (files: File[]) => void;
}

export function UploadDropZone({ onFilesSelected }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);

  function pickFiles(list: FileList | null) {
    if (!list || list.length === 0) return;
    onFilesSelected(Array.from(list));
  }

  function onDragOver(e: React.DragEvent) {
    e.preventDefault();
    setDragOver(true);
  }

  function onDragLeave(e: React.DragEvent) {
    e.preventDefault();
    setDragOver(false);
  }

  function onDrop(e: React.DragEvent) {
    e.preventDefault();
    setDragOver(false);
    pickFiles(e.dataTransfer.files);
  }

  return (
    <div
      className={`dropZone${dragOver ? " dropZoneActive" : ""}`}
      role="button"
      tabIndex={0}
      onClick={() => inputRef.current?.click()}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          inputRef.current?.click();
        }
      }}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      <input
        ref={inputRef}
        type="file"
        multiple
        className="dropZoneInput"
        onChange={(e) => {
          pickFiles(e.target.files);
          e.target.value = "";
        }}
      />
      <p className="dropZoneText">Drop files here or click to browse</p>
    </div>
  );
}
