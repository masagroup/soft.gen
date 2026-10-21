// Formats a folder of generated TypeScript files as soft.ts does after a
// generation (its "pretty" npm script): the unused import specifiers are
// removed (oxlint no-unused-vars, as soft.ts/ecore/scripts/remove-unused-imports.mjs),
// then every .ts file is formatted with prettier and the soft.ts options.
// The golden tests of lang/typescript run it on the output of the v1
// generator and on the expected files.
//
// Usage: node format.mjs <tools> <printWidth> <dir> [--keep-imports] [--lenient]
//
//   tools        a directory whose node_modules resolve prettier and oxlint
//                (soft.ts/ecore, or the npm workspace root)
//   printWidth   the prettier printWidth (120 for the soft.ts examples)
//   --keep-imports  only run prettier
//   --lenient       leave a file prettier can not parse as it is (the output of
//                   v1 may not parse)

import { spawnSync } from "node:child_process"
import { mkdtempSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from "node:fs"
import { createRequire } from "node:module"
import { tmpdir } from "node:os"
import { dirname, join, resolve } from "node:path"
import { pathToFileURL } from "node:url"

const [tools, printWidth, dir, ...flags] = process.argv.slice(2)
if (!tools || !printWidth || !dir) {
    console.error("usage: node format.mjs <tools> <printWidth> <dir> [--keep-imports]")
    process.exit(2)
}
const require = createRequire(join(resolve(tools), "package.json"))

function tsFiles(root) {
    const files = []
    for (const entry of readdirSync(root)) {
        const path = join(root, entry)
        if (statSync(path).isDirectory()) files.push(...tsFiles(path))
        else if (path.endsWith(".ts")) files.push(path)
    }
    return files
}

function unusedImports(root) {
    const oxlint = join(dirname(require.resolve("oxlint/package.json")), "bin", "oxlint")
    const tmp = mkdtempSync(join(tmpdir(), "unused-imports-"))
    try {
        const config = join(tmp, "oxlintrc.json")
        writeFileSync(
            config,
            JSON.stringify({
                plugins: ["typescript"],
                categories: { correctness: "off" },
                rules: { "no-unused-vars": ["error", { args: "none", caughtErrors: "none" }] }
            })
        )
        const result = spawnSync(process.execPath, [oxlint, "-c", config, "-f", "json", root], {
            encoding: "utf8",
            maxBuffer: 64 * 1024 * 1024
        })
        if (result.error) throw result.error
        const unused = new Map()
        for (const d of JSON.parse(result.stdout).diagnostics) {
            if (!/is imported but never used/.test(d.message)) continue
            const name = /'([^']+)'/.exec(d.message)[1]
            const file = resolve(d.filename)
            if (!unused.has(file)) unused.set(file, [])
            for (const l of d.labels) unused.get(file).push({ name, ...l.span })
        }
        return unused
    } finally {
        rmSync(tmp, { recursive: true, force: true })
    }
}

const importRe = /^import\s+(?!type\b)([\s\S]*?)\s+from\s+("[^"]+"|'[^']+')[ \t]*;?[ \t]*\r?\n/gm

function removeUnusedImports(root) {
    for (const [file, spans] of unusedImports(root)) {
        const text = readFileSync(file, "utf8")
        const bytes = Buffer.from(text, "utf8")
        const names = spans.map((s) => ({ name: s.name, offset: bytes.subarray(0, s.offset).toString("utf8").length }))
        const out = text.replace(importRe, (decl, clause, source, offset) => {
            const end = offset + decl.length
            const drop = new Set(names.filter((n) => n.offset >= offset && n.offset < end).map((n) => n.name))
            if (drop.size === 0) return decl
            const m = /^(?:([\w$]+)\s*,?\s*)?(?:\{([\s\S]*)\}|(\*\s+as\s+[\w$]+))?$/.exec(clause.trim())
            if (!m) throw new Error(`${file}: unsupported import clause '${clause}'`)
            const [, def, named, ns] = m
            const localName = (s) => s.split(/\s+/).pop()
            const keepDefault = def && !drop.has(def) ? def : null
            const keepNs = ns && !drop.has(localName(ns)) ? ns : null
            const keepNamed = (named ?? "")
                .split(",")
                .map((s) => s.trim())
                .filter((s) => s.length > 0 && !drop.has(localName(s)))
            const parts = [keepDefault, keepNs, keepNamed.length > 0 ? `{ ${keepNamed.join(", ")} }` : null].filter(Boolean)
            return parts.length > 0 ? `import ${parts.join(", ")} from ${source}\n` : ""
        })
        if (out !== text) writeFileSync(file, out)
    }
}

const root = resolve(dir)
if (!flags.includes("--keep-imports")) removeUnusedImports(root)
const prettierModule = await import(pathToFileURL(require.resolve("prettier")).href)
const prettier = prettierModule.format ? prettierModule : prettierModule.default
const options = { parser: "typescript", printWidth: Number(printWidth), tabWidth: 4, semi: false, trailingComma: "none" }
let failed = false
for (const file of tsFiles(root)) {
    const text = readFileSync(file, "utf8")
    try {
        const formatted = await prettier.format(text, options)
        if (formatted !== text) writeFileSync(file, formatted)
    } catch (e) {
        if (flags.includes("--lenient")) continue
        // the message of a syntax error, without the source prettier appends
        console.error(`${file}: ${String(e.message).split("\n")[0].slice(0, 300)}`)
        failed = true
    }
}
if (failed) process.exit(1)
