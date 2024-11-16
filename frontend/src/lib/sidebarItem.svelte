<script lang="ts" context="module">
  export type SidebarItemSelectEvent = CustomEvent<{
    group: string;
    version: string;
    name: string;
    namespaced: boolean;
  }>;
</script>

<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { Collapse } from "bootstrap";

  export let group: string;
  export let version: string;
  export let resources: { name: string; namespaced: boolean }[];

  let collapse: Collapse;
  let collapseEl: HTMLDivElement;

  const dispatch = createEventDispatcher();

  function onToggle() {
    collapse.toggle();
  }
  function onSelect(event: MouseEvent) {
    const target = event.target as HTMLElement;
    const name = target.dataset.name;
    const namespaced = target.dataset.namespaced === "true";

    dispatch("select", { group, version, name, namespaced });
  }

  onMount(() => {
    collapse = new Collapse(collapseEl, { toggle: false });
  });
</script>

<li>
  <button class="btn" on:click={onToggle}>
    {group}/{version}
  </button>
  <div class="collapse" bind:this={collapseEl}>
    <ul class="list-unstyled ms-4">
      {#each resources as res (res.name)}
        <a
          href={"#"}
          on:click={onSelect}
          data-name={res.name}
          data-namespaced={res.namespaced}
        >
          {res.name}
          {#if res.namespaced}N{/if}
        </a><br />
      {/each}
    </ul>
  </div>
</li>
