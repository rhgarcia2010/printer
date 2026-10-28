<script lang="ts">
  import type { book } from '../wailsjs/go/models'

  let {
    settings = $bindable(),
    papers,
    defaults,
  }: { settings: book.Settings; papers: book.Paper[]; defaults: book.Settings } = $props()

  const binding = $derived(settings.sheetsPerSignature === 0 ? 'booklet' : 'signatures')

  function setBinding(kind: 'booklet' | 'signatures') {
    settings.sheetsPerSignature = kind === 'booklet' ? 0 : 4
  }

  function reset() {
    settings = { ...defaults }
  }
</script>

<div class="settings">
  <section>
    <label for="paper">Paper</label>
    <select id="paper" bind:value={settings.paperId}>
      {#each papers as p}
        <option value={p.id}>{p.name}</option>
      {/each}
    </select>
    <p class="hint">Each sheet is folded in half, giving 4 book pages.</p>
  </section>

  <section>
    <span class="label">Binding</span>
    <div class="segmented">
      <button class:active={binding === 'booklet'} onclick={() => setBinding('booklet')}>
        Stapled booklet
      </button>
      <button class:active={binding === 'signatures'} onclick={() => setBinding('signatures')}>
        Signatures
      </button>
    </div>
    {#if binding === 'signatures'}
      <div class="row">
        <label for="per-sig">Sheets per signature</label>
        <input id="per-sig" type="number" min="1" max="16" bind:value={settings.sheetsPerSignature} />
      </div>
      <p class="hint">Fold each group separately, stack them in order, then sew or glue the spine.</p>
    {:else}
      <p class="hint">All sheets nest inside each other and are stapled on the fold.</p>
    {/if}
  </section>

  <section>
    <label for="blanks">Extra blank pages go</label>
    <select id="blanks" bind:value={settings.blanks}>
      <option value="before-back-cover">Before the back cover</option>
      <option value="end">At the very end</option>
    </select>
  </section>

  <section>
    <span class="label">Page scaling</span>
    <div class="segmented">
      <button class:active={settings.scaling === 'fit'} onclick={() => (settings.scaling = 'fit')}>Fit to half sheet</button>
      <button class:active={settings.scaling === 'actual'} onclick={() => (settings.scaling = 'actual')}>Actual size</button>
    </div>
  </section>

  <section>
    <span class="label">Two-sided printing</span>
    <div class="segmented">
      <button class:active={settings.duplex === 'manual'} onclick={() => (settings.duplex = 'manual')}>Manual</button>
      <button class:active={settings.duplex === 'auto'} onclick={() => (settings.duplex = 'auto')}>Automatic</button>
    </div>
    <p class="hint">
      {#if settings.duplex === 'auto'}
        The printer flips each sheet on the short edge. Check your printer supports it for this paper size.
      {:else}
        Print all fronts, put the stack back in, then print the backs.
      {/if}
    </p>
  </section>

  <section>
    <span class="label">Margins</span>
    <div class="row">
      <label for="pm">Printer can't print within</label>
      <div class="unit"><input id="pm" type="number" min="0" max="20" step="0.5" bind:value={settings.printerMarginMm} /><span>mm</span></div>
    </div>
    <div class="row">
      <label for="sm">Warn when content is within</label>
      <div class="unit"><input id="sm" type="number" min="0" max="30" step="0.5" bind:value={settings.safeMarginMm} /><span>mm</span></div>
    </div>
    <div class="row">
      <label for="creep" title="Thick booklets: inner sheets stick out at the edge after folding. This shifts inner pages toward the fold to compensate.">
        Creep (innermost sheet)
      </label>
      <div class="unit"><input id="creep" type="number" min="0" max="10" step="0.1" bind:value={settings.creepMm} /><span>mm</span></div>
    </div>
  </section>

  <section>
    <label class="check">
      <input type="checkbox" bind:checked={settings.foldMarks} />
      Print fold marks
    </label>
  </section>

  <button class="link" onclick={reset}>Reset to defaults</button>
</div>

<style>
  .settings { display: flex; flex-direction: column; gap: 18px; }
  section { display: flex; flex-direction: column; gap: 6px; }
  label, .label { font-weight: 600; font-size: 13px; }
  .row { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
  .row label { font-weight: 400; color: var(--muted-text); }
  .row input { width: 64px; }
  .unit { display: flex; align-items: center; gap: 4px; color: var(--muted-text); }
  .hint { margin: 0; font-size: 12px; color: var(--muted-text); line-height: 1.4; }
  .segmented { display: flex; border: 1px solid var(--border); border-radius: 6px; overflow: hidden; }
  .segmented button {
    flex: 1; border: none; border-radius: 0; padding: 6px 8px; font-size: 13px; background: var(--surface);
  }
  .segmented button + button { border-left: 1px solid var(--border); }
  .segmented button.active { background: var(--accent-soft); color: var(--accent); font-weight: 600; }
  .check { display: flex; align-items: center; gap: 8px; font-weight: 400; }
  .check input { accent-color: var(--accent); }
  .link {
    border: none; background: none; padding: 0; color: var(--muted-text);
    text-decoration: underline; font-size: 12px; align-self: flex-start;
  }
</style>
