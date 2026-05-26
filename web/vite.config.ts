import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// During development the UI runs on Vite's dev server and proxies /api calls
// to the Go backend. In production the built dist/ is embedded into the Go
// binary (see web/embed.go).
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
