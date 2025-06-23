/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  // Ajoutez d'autres variables d'environnement ici
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
} 