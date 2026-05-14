declare module '*.svg' {
  const content: string;
  export default content;
}

declare module '*.css';

interface ImportMetaEnv {
  readonly VITE_BACKEND_URL: string;
  readonly VITE_GOOGLE_LOGIN: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}