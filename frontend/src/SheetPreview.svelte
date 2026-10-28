<script lang="ts">
  import type { book, job } from '../wailsjs/go/models'

  let {
    side,
    sheet,
    settings,
    showMargins,
    highlighted,
  }: {
    side: job.SidePreview
    sheet: book.Size
    settings: book.Settings
    showMargins: boolean
    highlighted: boolean
  } = $props()

  const MM = 72 / 25.4
  // Margin band thickness as a percentage of each sheet dimension.
  const bandX = $derived(((settings.printerMarginMm * MM) / sheet.width) * 100)
  const bandY = $derived(((settings.printerMarginMm * MM) / sheet.height) * 100)
</script>

<figure class:highlighted>
  <div class="paper" style:aspect-ratio="{sheet.width} / {sheet.height}">
    <img src={side.image} alt="Sheet {side.sheet} {side.back ? 'back' : 'front'}" />
    {#if showMargins}
      <div class="band" style:top="0" style:left="0" style:right="0" style:height="{bandY}%"></div>
      <div class="band" style:bottom="0" style:left="0" style:right="0" style:height="{bandY}%"></div>
      <div class="band" style:top="{bandY}%" style:bottom="{bandY}%" style:left="0" style:width="{bandX}%"></div>
      <div class="band" style:top="{bandY}%" style:bottom="{bandY}%" style:right="0" style:width="{bandX}%"></div>
    {/if}
    <div class="fold"></div>
  </div>
  <figcaption>
    <span class="side">{side.back ? 'Back' : 'Front'}</span>
    <span class="pages">
      <span class:blank={side.left === 'blank'}>{side.left === 'blank' ? 'blank' : `p. ${side.left}`}</span>
      <span class="sep">|</span>
      <span class:blank={side.right === 'blank'}>{side.right === 'blank' ? 'blank' : `p. ${side.right}`}</span>
    </span>
  </figcaption>
</figure>

<style>
  figure { margin: 0; display: flex; flex-direction: column; gap: 6px; min-width: 0; }
  .paper {
    position: relative;
    background: white;
    box-shadow: var(--shadow);
    border-radius: 2px;
    overflow: hidden;
    outline: 3px solid transparent;
    transition: outline-color 0.15s;
  }
  figure.highlighted .paper { outline-color: var(--accent); }
  img { display: block; width: 100%; height: 100%; }
  .band {
    position: absolute;
    background: repeating-linear-gradient(45deg, var(--danger-band) 0 4px, transparent 4px 8px);
    pointer-events: none;
  }
  .fold {
    position: absolute; top: 0; bottom: 0; left: 50%;
    border-left: 1px dashed rgba(0, 0, 0, 0.25);
    pointer-events: none;
  }
  figcaption {
    display: flex; justify-content: space-between; font-size: 12px; color: var(--muted-text);
  }
  .side { font-weight: 600; }
  .pages { display: flex; gap: 6px; font-variant-numeric: tabular-nums; }
  .sep { opacity: 0.5; }
  .blank { font-style: italic; opacity: 0.7; }
</style>
