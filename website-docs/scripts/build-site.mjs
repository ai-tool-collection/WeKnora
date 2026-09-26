import { spawnSync } from 'node:child_process'
import { cpSync, rmSync, mkdirSync, writeFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
for (const script of ['check:links', 'build:docs']) {
  const result = spawnSync('npm', ['run', script], { cwd: root, stdio: 'inherit', shell: process.platform === 'win32' })
  if (result.status !== 0) process.exit(result.status || 1)
}

const destination = resolve(root, 'static-site')
rmSync(destination, { recursive: true, force: true })
mkdirSync(destination, { recursive: true })
cpSync(resolve(root, '.vitepress/dist'), resolve(destination, 'docs'), { recursive: true })
writeFileSync(resolve(destination, 'index.html'), '<!doctype html><html lang="en"><head><meta charset="utf-8"><meta http-equiv="refresh" content="0;url=/docs/"><title>Knowledge Hub</title></head><body><a href="/docs/">Knowledge Hub documentation</a></body></html>\n')
writeFileSync(resolve(destination, '404.html'), '<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Not found · Knowledge Hub</title></head><body><h1>Not found</h1><a href="/docs/">Documentation</a></body></html>\n')
console.log('Knowledge Hub site ready in static-site/')
