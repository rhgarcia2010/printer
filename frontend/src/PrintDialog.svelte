<script lang="ts">
  import { untrack } from 'svelte'
  import { main, type book } from '../wailsjs/go/models'
  import { Print } from '../wailsjs/go/main/App'

  let {
    settings,
    options,
    sheetCount,
    onclose,
    onmessage,
  }: {
    settings: book.Settings
    options: main.Options
    sheetCount: number
    onclose: () => void
    onmessage: (kind: 'ok' | 'error', text: string) => void
  } = $props()

  // The dialog starts from the remembered choices and then owns them.
  const saved = untrack(() => options)
  let printer = $state(
    saved.config.printer || saved.printers?.find((p) => p.default)?.name || saved.printers?.[0]?.name || '',
  )
  let reverseBacks = $state(saved.config.reverseBacks)
  let rotateBacks = $state(saved.config.rotateBacks)
  let busy = $state(false)
  let frontsDone = $state(false)

  const manual = $derived(settings.duplex === 'manual')

  async function send(pass: 'all' | 'fronts' | 'backs') {
    busy = true
    try {
      const viewerOpened = await Print(
        main.PrintRequest.createFrom({ settings, printer, pass, reverseBacks, rotateBacks }),
      )
      if (viewerOpened) {
        onmessage('ok', 'Opened in your PDF viewer. Print it from there with scaling set to "Actual size".')
      } else {
        onmessage('ok', pass === 'fronts' ? 'Fronts sent to the printer.' : 'Sent to the printer.')
      }
      if (pass === 'fronts') frontsDone = true
      else onclose()
    } catch (e) {
      onmessage('error', String(e))
    } finally {
      busy = false
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="backdrop" onclick={onclose}>
  <div class="dialog" role="dialog" tabindex="-1" aria-modal="true" aria-labelledby="print-title" onclick={(e) => e.stopPropagation()}>
    <h2 id="print-title">Print {sheetCount} sheet{sheetCount === 1 ? '' : 's'}</h2>

    {#if options.directPrint}
      <label for="printer">Printer</label>
      <select id="printer" bind:value={printer}>
        {#each options.printers ?? [] as p}
          <option value={p.name}>{p.name}{p.default ? ' (default)' : ''}</option>
        {/each}
      </select>
      {#if !options.printers?.length}
        <p class="note">No printers found. Add your printer in the system settings first.</p>
      {/if}
    {:else}
      <p class="note">
        The PDF opens in your PDF viewer. Print it from there with scaling set to <b>Actual size</b> and paper set to
        the size chosen here.
      </p>
    {/if}

    {#if !manual}
      <p>The printer prints both sides and flips each sheet on the <b>short edge</b>.</p>
      <div class="actions">
        <button onclick={onclose}>Cancel</button>
        <button class="primary" disabled={busy} onclick={() => send('all')}>Print</button>
      </div>
    {:else}
      <ol class="steps">
        <li class:done={frontsDone}>
          <div>
            <b>Print the fronts.</b>
            <span>One side of every sheet.</span>
          </div>
          <button class:primary={!frontsDone} disabled={busy} onclick={() => send('fronts')}>
            {frontsDone ? 'Print fronts again' : 'Print fronts'}
          </button>
        </li>
        <li>
          <div>
            <b>Put the stack back in the printer</b>
            <span>with the printed side turned over, then print the backs.</span>
            <label class="check"><input type="checkbox" bind:checked={reverseBacks} /> Print backs in reverse order</label>
            <label class="check"><input type="checkbox" bind:checked={rotateBacks} /> Rotate backs upside down</label>
          </div>
          <button class:primary={frontsDone} disabled={busy} onclick={() => send('backs')}>Print backs</button>
        </li>
      </ol>
      <p class="note">
        Feeding differs between printers. Try the first time with two scrap sheets and adjust the two options above until
        the page numbers line up; they are remembered.
      </p>
      <div class="actions">
        <button onclick={onclose}>Done</button>
      </div>
    {/if}
  </div>
</div>

<style>
  .backdrop {
    position: fixed; inset: 0; background: rgba(0, 0, 0, 0.35);
    display: grid; place-items: center; z-index: 10;
  }
  .dialog {
    background: var(--panel); border-radius: 12px; box-shadow: var(--shadow);
    width: min(520px, calc(100vw - 32px)); padding: 22px 24px; display: flex; flex-direction: column; gap: 10px;
  }
  h2 { margin: 0 0 4px; font-size: 18px; }
  label { font-weight: 600; font-size: 13px; }
  p { margin: 0; line-height: 1.45; }
  .note { font-size: 12px; color: var(--muted-text); }
  .steps { margin: 6px 0; padding: 0; list-style: none; counter-reset: step; display: flex; flex-direction: column; gap: 10px; }
  .steps li {
    counter-increment: step; display: flex; gap: 12px; align-items: flex-start; justify-content: space-between;
    padding: 12px; border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface);
  }
  .steps li > div { display: flex; flex-direction: column; gap: 4px; }
  .steps li > div::before { content: 'Step ' counter(step); font-size: 11px; color: var(--muted-text); text-transform: uppercase; letter-spacing: 0.05em; }
  .steps li.done > div::before { content: 'Step ' counter(step) ' ✓'; color: var(--accent); }
  .steps span { color: var(--muted-text); font-size: 13px; }
  .steps button { white-space: nowrap; }
  .check { display: flex; gap: 6px; align-items: center; font-weight: 400; margin-top: 4px; }
  .check input { accent-color: var(--accent); }
  .actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 6px; }
</style>
