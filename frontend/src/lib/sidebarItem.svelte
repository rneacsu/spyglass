<script lang="ts" module>
  export type SidebarItemSelectEvent = {
    group: string;
    version: string;
    resource: string;
    namespaced: boolean;
  };
</script>

<script lang="ts">
  import { onMount } from "svelte";
  import { Collapse } from "bootstrap";

  let {
    group = "",
    version = "",
    resources = [],
    select = () => {},
  }: {
    group: string;
    version: string;
    resources: { name: string; namespaced: boolean }[];
    select: (event: SidebarItemSelectEvent) => void;
  } = $props();

  let collapse: Collapse;
  let collapseEl: HTMLDivElement;

  function onToggle() {
    collapse.toggle();
  }

  function onSelect(event: MouseEvent) {
    const target = event.target as HTMLElement;
    const resource = target.dataset.resource || "";
    const namespaced = target.dataset.namespaced === "true";

    select({ group, version, resource, namespaced });
  }

  onMount(() => {
    collapse = new Collapse(collapseEl, { toggle: false });
  });
</script>

<li>
  <button class="btn" onclick={onToggle}>
    {group}/{version}
  </button>
  <div class="collapse" bind:this={collapseEl}>
    <ul class="list-unstyled ms-4">
      {#each resources as res (res.name)}
        <a
          href={"#"}
          onclick={onSelect}
          data-resource={res.name}
          data-namespaced={res.namespaced}
        >
          {res.name}
          {#if res.namespaced}N{/if}
        </a><br />
      {/each}
    </ul>
  </div>
</li>
