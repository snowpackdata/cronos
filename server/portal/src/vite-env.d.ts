/// <reference types="vite/client" />

/**
 * Extends the ImportMeta interface for Vite environment variables
 * This provides TypeScript type safety for import.meta.env
 */
interface ImportMetaEnv {
  readonly VITE_APP_TITLE: string
  // more env variables...
}

interface ImportMeta {
  readonly env: ImportMetaEnv
} 