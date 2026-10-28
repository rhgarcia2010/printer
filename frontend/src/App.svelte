<script lang="ts">
  import { onMount } from 'svelte'
  import { Analyze, ChooseFile, OpenFile, OpenInViewer, Options, Reload, SavePDF, StartupFile } from '../wailsjs/go/main/App'
  import { OnFileDrop } from '../wailsjs/runtime/runtime'
  import type { book, job, main } from '../wailsjs/go/models'
  import Settings from './Settings.svelte'
  import SheetPreview from './SheetPreview.svelte'
  import PrintDialog from './PrintDialog.svelte'

  let options = $state<main.Options | null>(null)
  let settings = $state<book.Settings | null>(null)
  let info = $state<job.Info | null>(null)
  let result = $state<job.Result | null>(null)
  let loading = $state(false)
  let analyzing = $state(false)
  let error = $state('')
  let toast = $state<{ kind: 'ok' | 'error'; text: string } | null>(null)
  let showMargins = $state(true)
  let printing = $state(false)
  let selectedIssue = $state<number | null>(null)

  onMount(async () => {
    options = await Options()
    settings = { ...options.config.settings }
    OnFileDrop((_x, _y, paths) => paths[0] && open(() => OpenFile(paths[0])), false)
    const initial = await StartupFile()
    if (initial) open(() => OpenFile(initial))
  })

  async function open(load: () => Promise<job.Info>) {
    loading = true
    error = ''
    try {
      const loaded = await load()
      if (loaded) {
        info = loaded
        result = null
        selectedIssue = null
      }
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  // Re-run the layout and checks whenever the file or a setting changes.
  let seq = 0
  $effect(() => {
    if (!info || !settings) return
    const snapshot = $state.snapshot(settings) as book.Settings
    void info.path
    const mine = ++seq
    const timer = setTimeout(async () => {
      analyzing = true
      try {
        const res = await Analyze(snapshot)
        if (mine === seq) {
          result = res
          error = ''
        }
      } catch (e) {
        if (mine === seq) error = String(e)
      } finally {
        if (mine === seq) analyzing = false
      }
    }, 250)
    return () => clearTimeout(timer)
  })

  function notify(kind: 'ok' | 'error', text: string) {
    toast = { kind, text }
    setTimeout(() => (toast = null), kind === 'error' ? 8000 : 4000)
  }

  async function save() {
    try {
      const path = await SavePDF($state.snapshot(settings!) as book.Settings)
      if (path) notify('ok', `Saved ${path}`)
    } catch (e) {
      notify('error', String(e))
    }
  }

  async function viewer() {
    try {
      await OpenInViewer($state.snapshot(settings!) as book.Settings)
    } catch (e) {
      notify('error', String(e))
    }
  }

  const sheets = $derived.by(() => {
    const out: { number: number; signature: number; front: job.SidePreview; back: job.SidePreview }[] = []
    const sides = result?.sides ?? []
    for (let i = 0; i + 1 < sides.length; i += 2) {
      out.push({ number: sides[i].sheet, signature: sides[i].signature, front: sides[i], back: sides[i + 1] })
    }
    return out
  })

  const highlightPages = $derived(
    new Set((selectedIssue !== null ? (result?.issues[selectedIssue]?.pages ?? []) : []).map(String)),
  )
  const isHighlighted = (s: job.SidePreview) => highlightPages.has(s.left) || highlightPages.has(s.right)

  const counts = $derived.by(() => {
    const c = { error: 0, warning: 0, info: 0 }
    for (const i of result?.issues ?? []) c[i.severity as keyof typeof c]++
    return c
  })

  const inches = (pt: number) => +(pt / 72).toFixed(2)
  const mm = (pt: number) => Math.round((pt / 72) * 25.4)
</script>

<div class="layout">
  <aside>
    <header>
      <h1>Book Printer</h1>
    </header>

    <div class="file">
      {#if info}
        <div class="file-name" title={info.path}>{info.name}</div>
        <div class="file-meta">
          {info.pageCount} pages · {inches(info.pageSize.width)} × {inches(info.pageSize.height)} in
          ({mm(info.pageSize.width)} × {mm(info.pageSize.height)} mm)
        </div>
        <div class="file-actions">
          <button onclick={() => open(ChooseFile)} disabled={loading}>Open other…</button>
          <button onclick={() => open(Reload)} disabled={loading} title="Read the file again after re-exporting it from Canva">
            Reload
          </button>
        </div>
      {:else}
        <button class="primary" onclick={() => open(ChooseFile)} disabled={loading}>Open PDF…</button>
        {#if options?.config.lastFile}
          <button class="recent" onclick={() => open(() => OpenFile(options!.config.lastFile))} disabled={loading}>
            Reopen {options.config.lastFile.split(/[\\/]/).pop()}
          </button>
        {/if}
      {/if}
    </div>

    {#if settings && options}
      <Settings bind:settings papers={options.papers} defaults={options.defaults} />
    {/if}
  </aside>

  <main>
    {#if !info}
      <div class="empty">
        <div class="drop">
          <div class="drop-icon" aria-hidden="true">
            <svg viewBox="0 0 64 48" width="72"><path d="M4 6 L32 10 L60 6 L60 42 L32 46 L4 42 Z" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"/><path d="M32 10 V46" stroke="currentColor" stroke-width="2" stroke-dasharray="3 3"/></svg>
          </div>
          <h2>{loading ? 'Opening…' : 'Drop your book PDF here'}</h2>
          <p>Export your book from Canva as a PDF with one page per book page, in reading order.</p>
          <button class="primary" onclick={() => open(ChooseFile)} disabled={loading}>Choose a PDF…</button>
          {#if error}<p class="err">{error}</p>{/if}
        </div>
      </div>
    {:else}
      <div class="toolbar">
        <div class="summary">
          {#if result}
            <b>{result.plan.sheets.length} sheet{result.plan.sheets.length === 1 ? '' : 's'}</b>
            <span>{result.plan.bookPages} pages</span>
            {#if result.plan.signatures > 1}<span>{result.plan.signatures} signatures</span>{/if}
          {/if}
          {#if analyzing || loading}<span class="spinner" aria-label="Updating"></span>{/if}
        </div>
        <label class="toggle"><input type="checkbox" bind:checked={showMargins} /> Show unprintable margin</label>
        <div class="actions">
          <button onclick={viewer} disabled={!result}>Open in viewer</button>
          <button onclick={save} disabled={!result}>Save PDF…</button>
          <button class="primary" onclick={() => (printing = true)} disabled={!result || counts.error > 0}>Print…</button>
        </div>
      </div>

      {#if error}<div class="banner err">{error}</div>{/if}

      <div class="content">
        {#if result && settings}
          <section class="issues" aria-label="Checks">
            {#each result.issues as issue, i}
              <button
                class="issue {issue.severity}"
                class:selected={selectedIssue === i}
                onclick={() => (selectedIssue = selectedIssue === i ? null : i)}
                disabled={!issue.pages?.length}
              >
                <span class="dot" aria-hidden="true"></span>
                <span class="text">
                  <b>{issue.title}</b>
                  <span>{issue.detail}</span>
                  {#if issue.pages?.length}
                    <em>{selectedIssue === i ? 'Highlighted below. Click again to clear.' : 'Click to highlight the affected sheets.'}</em>
                  {/if}
                </span>
              </button>
            {/each}
          </section>

          <section class="sheets">
            {#each sheets as sh (sh.number)}
              <div class="sheet">
                <h3>
                  Sheet {sh.number}
                  {#if result.plan.signatures > 1}<span>· signature {sh.signature}</span>{/if}
                </h3>
                <div class="sides">
                  <SheetPreview side={sh.front} sheet={result.sheet} {settings} {showMargins} highlighted={isHighlighted(sh.front)} />
                  <SheetPreview side={sh.back} sheet={result.sheet} {settings} {showMargins} highlighted={isHighlighted(sh.back)} />
                </div>
              </div>
            {/each}
          </section>
        {:else if !error}
          <div class="placeholder">Laying out sheets…</div>
        {/if}
      </div>
    {/if}
  </main>
</div>

{#if printing && options && settings && result}
  <PrintDialog
    settings={$state.snapshot(settings) as book.Settings}
    {options}
    sheetCount={result.plan.sheets.length}
    onclose={() => (printing = false)}
    onmessage={notify}
  />
{/if}

{#if toast}
  <div class="toast {toast.kind}" role="status">{toast.text}</div>
{/if}

<style>
  .layout { display: grid; grid-template-columns: 300px 1fr; height: 100vh; }
  aside {
    background: var(--panel); border-right: 1px solid var(--border);
    padding: 18px 20px 28px; overflow-y: auto; display: flex; flex-direction: column; gap: 20px;
  }
  header h1 { margin: 0; font-size: 17px; letter-spacing: 0.01em; }
  .file {
    display: flex; flex-direction: column; gap: 6px; padding: 12px;
    background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius);
  }
  .file-name { font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .file-meta { font-size: 12px; color: var(--muted-text); }
  .file-actions { display: flex; gap: 6px; margin-top: 4px; }
  .file-actions button { flex: 1; padding: 5px 8px; font-size: 13px; }
  .recent { font-size: 12px; border: none; background: none; color: var(--muted-text); text-decoration: underline; padding: 2px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

  main { display: flex; flex-direction: column; min-width: 0; min-height: 0; }
  .toolbar {
    display: flex; align-items: center; gap: 16px; padding: 12px 22px;
    border-bottom: 1px solid var(--border); background: var(--panel); flex-wrap: wrap;
  }
  .summary { display: flex; align-items: center; gap: 12px; flex: 1; color: var(--muted-text); }
  .summary b { color: var(--text); }
  .toggle { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--muted-text); }
  .toggle input { accent-color: var(--accent); }
  .actions { display: flex; gap: 8px; }

  .content { overflow-y: auto; padding: 18px 22px 40px; display: flex; flex-direction: column; gap: 22px; }
  .issues { display: flex; flex-direction: column; gap: 8px; max-width: 980px; }
  .issue {
    display: flex; gap: 10px; text-align: left; align-items: flex-start; padding: 10px 14px;
    border-radius: var(--radius); border: 1px solid transparent; line-height: 1.45;
  }
  .issue:disabled { cursor: default; opacity: 1; }
  .issue .dot { width: 8px; height: 8px; border-radius: 50%; margin-top: 6px; flex-shrink: 0; }
  .issue .text { display: flex; flex-direction: column; gap: 2px; }
  .issue .text span { color: var(--text); opacity: 0.85; }
  .issue em { font-size: 12px; color: var(--muted-text); font-style: normal; }
  .issue.error { background: var(--error-soft); }
  .issue.error .dot { background: var(--error); }
  .issue.warning { background: var(--warning-soft); }
  .issue.warning .dot { background: var(--warning); }
  .issue.info { background: var(--info-soft); }
  .issue.info .dot { background: var(--info); }
  .issue.selected { border-color: var(--accent); }

  .sheets { display: grid; grid-template-columns: repeat(auto-fill, minmax(560px, 1fr)); gap: 22px 28px; }
  .sheet h3 { margin: 0 0 8px; font-size: 13px; }
  .sheet h3 span { font-weight: 400; color: var(--muted-text); }
  .sides { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

  .empty { flex: 1; display: grid; place-items: center; padding: 24px; }
  .drop {
    text-align: center; max-width: 440px; padding: 44px 40px; border: 2px dashed var(--border);
    border-radius: 16px; display: flex; flex-direction: column; align-items: center; gap: 10px;
  }
  .drop-icon { color: var(--accent); }
  .drop h2 { margin: 6px 0 0; font-size: 20px; }
  .drop p { margin: 0 0 10px; color: var(--muted-text); line-height: 1.5; }
  .err { color: var(--error); }
  .banner { margin: 14px 22px 0; padding: 10px 14px; border-radius: var(--radius); background: var(--error-soft); }
  .placeholder { color: var(--muted-text); padding: 40px; text-align: center; }

  .spinner {
    width: 14px; height: 14px; border: 2px solid var(--border); border-top-color: var(--accent);
    border-radius: 50%; animation: spin 0.7s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  .toast {
    position: fixed; bottom: 20px; left: 50%; transform: translateX(-50%); z-index: 20;
    padding: 10px 18px; border-radius: var(--radius); box-shadow: var(--shadow);
    background: var(--surface); border: 1px solid var(--border); max-width: 80vw;
  }
  .toast.error { border-color: var(--error); color: var(--error); }
</style>
