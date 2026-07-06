 import { defineConfig } from 'vite';
 import react from '@vitejs/plugin-react';
 import fs from 'fs';
 import path from 'path';
 
 export default defineConfig({
   plugins: [
     react(),
     {
       name: 'copy-readme',
       closeBundle() {
         // __dirname = frontend/ ; project root is one level up
         const root = path.resolve(__dirname, '..');
         const src = path.join(root, 'README.md');
         const dest = path.join(root, 'docs', 'README.md');
         if (fs.existsSync(src)) {
           fs.copyFileSync(src, dest);
           console.log('[copy-readme] README.md copied to docs/');
         }
       }
     }
   ],
   base: './',
   build: {
     outDir: '../docs',
     emptyOutDir: true,
   },
 });
