<script lang="ts" module>
  export type SidebarItemConfig = {
    text: string;
    data: any | null;
    active: boolean;
    items: SidebarItemConfig[];
  };
</script>

<script lang="ts">
  import { onMount } from "svelte";
  import { Collapse } from "bootstrap";
  import Self from "./sidebarItem.svelte";

  let {
    config = { text: "", data: null, active: false, items: [] },
    select = () => {},
  }: {
    config: SidebarItemConfig;
    select: (event: any) => void;
  } = $props();

  let collapse: Collapse | null = null;
  let collapseEl: HTMLDivElement;

  function onToggle() {
    if (!config.active) {
      collapse?.toggle();
    }
  }

  onMount(() => {
    if (collapseEl) {
      collapse = new Collapse(collapseEl, { toggle: config.active });
    }
  });
</script>

<li>
  {#if config.items.length > 0}
    <button
      class="btn btn-small text-start text-truncate w-100 d-inline-block rounded-end-0"
      class:active={config.active}
      type="button"
      aria-expanded="false"
      onclick={onToggle}
    >
      {config.text}
    </button>
    <div class="collapse ps-3" bind:this={collapseEl}>
      <ul class="list-unstyled">
        {#each config.items as subItem (subItem.text)}
          <Self config={subItem} {select} />
        {/each}
      </ul>
    </div>
  {:else}
    <button
      class="btn text-start text-truncate w-100 d-inline-block rounded-end-0"
      class:active={config.active}
      onclick={() => select(config.data)}
    >
      {config.text}
    </button>
  {/if}
</li>

<style lang="scss">
  button {
    border: none;

    &:hover {
      background-color: var(--bs-tertiary-bg);
    }

    &.active {
      background-color: var(--bs-secondary-bg);
    }
  }
</style>
