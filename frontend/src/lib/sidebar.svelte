<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import SidebarItem, {
    type SidebarItemSelectEvent,
  } from "./sidebarItem.svelte";
  import client from "./grpc/client";
  import { ShowAlert } from "./alerts.svelte";
  import { Refresher } from "./grpc/refresher";
  import { SvelteMap } from "svelte/reactivity";

  let {
    context = "",
    group = $bindable(""),
    version = $bindable(""),
    resource = $bindable(""),
    namespaced = $bindable(false),
  } = $props();

  let apis: SvelteMap<
    { group: string; version: string },
    { name: string; namespaced: boolean }[]
  > = new SvelteMap();

  let isLoadingSidebar: boolean = $state(false);

  let sidebarRefresher: Refresher | null = null;

  $effect(() => {
    context !== null &&
      untrack(() => {
        onParamsChange();
      });
  });

  async function loadSidebar(signal: AbortSignal) {
    if (!context) {
      return;
    }

    ShowAlert("info", "Loading sidebar...");

    const res = await (
      await client
    ).discover({ context: context }, { signal: signal });

    apis.clear();

    const apiKeys = Object.keys(res.apis).sort();

    for (const gv of apiKeys) {
      const api = res.apis[gv];
      apis.set(
        { group: api.group, version: api.version },
        api.resources
          .map((r) => {
            return { name: r.name, namespaced: r.namespaced };
          })
          .sort((a, b) => a.name.localeCompare(b.name)),
      );
    }
  }

  function onParamsChange() {
    apis.clear();

    (async () => {
      isLoadingSidebar = true;
      await sidebarRefresher?.refresh();
      isLoadingSidebar = false;
    })();
  }

  function onSelect(event: SidebarItemSelectEvent) {
    ({ group, version, resource, namespaced } = event);
  }

  onMount(() => {
    sidebarRefresher = new Refresher({
      refresh: loadSidebar,
      onError: (e) => {
        ShowAlert("error", e.message);
      },
    });
  });

  onDestroy(() => {
    sidebarRefresher?.abort();
  });
</script>

<div>
  <ul class="list-unstyled">
    {#if isLoadingSidebar}
      <li>Loading...</li>
    {:else if apis}
      {#if apis.size === 0}
        <li>No resources found</li>
      {:else}
        {#each apis.entries() as [gv, resources] (gv.group + gv.version)}
          <SidebarItem
            group={gv.group}
            version={gv.version}
            {resources}
            select={onSelect}
          />
        {/each}
      {/if}
    {/if}
  </ul>
</div>
