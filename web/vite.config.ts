import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The build output is embedded into the Go binary via `//go:embed dist` in
// internal/server, so it must land in internal/server/dist.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../internal/server/dist",
    emptyOutDir: true,
    // Stable (unhashed) filenames keep the committed, embedded build from
    // producing churn in git on every rebuild.
    rollupOptions: {
      output: {
        entryFileNames: "assets/app.js",
        chunkFileNames: "assets/[name].js",
        assetFileNames: "assets/app.[ext]",
      },
    },
  },
  server: {
    // During `npm run dev`, proxy API calls to a running `tagbackup serve`.
    proxy: {
      "/api": "http://127.0.0.1:3000",
    },
  },
});
