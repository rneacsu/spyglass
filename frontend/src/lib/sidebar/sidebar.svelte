<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import SidebarItem, { type SidebarItemConfig } from "./sidebarItem.svelte";
  import client from "../grpc/client";
  import { ShowAlert } from "../alerts.svelte";
  import { Refresher } from "../grpc/refresher";
  import { translateResource } from "../translator";
  import { structure, hidden } from "./config";

  let {
    context = "",
    group = $bindable(""),
    version = $bindable(""),
    resource = $bindable(""),
    namespaced = $bindable(false),
  } = $props();

  let items: SidebarItemConfig[] = $state([]);
  let isLoadingSidebar: boolean = $state(false);

  let sidebarRefresher: Refresher | null = null;

  $effect(() => {
    context !== null &&
      untrack(() => {
        onParamsChange();
      });
  });

  function gvToKey(group: string, version: string) {
    return group + "/" + version;
  }

  function gvrToKey(group: string, version: string, resource: string) {
    return gvToKey(group, version) + "::" + resource;
  }

  function keyToGv(key: string) {
    const [group, version] = key.split("/");
    return { group, version };
  }

  function isActive(
    itemGroup: string,
    itemVersion: string,
    itemResource: string,
  ) {
    return (
      itemGroup === group &&
      itemVersion === version &&
      itemResource === resource
    );
  }

  async function loadSidebar(signal: AbortSignal) {
    if (!context) {
      return;
    }

    // ShowAlert("info", "Loading sidebar...");

    const discover = await (
      await client
    ).discover({ context: context }, { signal: signal });

    items = [];

    const apisGrouped = new Map<string, Map<string, { namespaced: boolean }>>();
    const apisFlattened = new Map<string, { namespaced: boolean }>();

    for (const gv of Object.keys(discover.apis).sort()) {
      const api = discover.apis[gv];
      const apiGroup = new Map<string, { namespaced: boolean }>();

      api.resources.sort((a, b) => a.name.localeCompare(b.name));
      for (const res of api.resources) {
        apiGroup.set(res.name, { namespaced: res.namespaced });
        apisFlattened.set(gvrToKey(api.group, api.version, res.name), {
          namespaced: res.namespaced,
        });
      }

      apisGrouped.set(gvToKey(api.group, api.version), apiGroup);
    }

    for (const [category, categoryItems] of Object.entries(structure)) {
      const subItems: SidebarItemConfig[] = [];
      for (const categoryItem of categoryItems) {
        const gvKey = gvToKey(categoryItem.group, categoryItem.version);
        const gvrKey = gvrToKey(
          categoryItem.group,
          categoryItem.version,
          categoryItem.resource,
        );
        const apiItems = apisFlattened.get(gvrKey);

        if (apiItems) {
          const data = {
            group: categoryItem.group,
            version: categoryItem.version,
            resource: categoryItem.resource,
            namespaced: apiItems.namespaced,
          };
          subItems.push({
            text: translateResource(gvrKey),
            data,
            items: [],
            active: isActive(data.group, data.version, data.resource),
          });
          apisGrouped.get(gvKey)?.delete(categoryItem.resource);
          apisFlattened.delete(gvrKey);
        }
      }

      if (subItems.length > 0) {
        items.push({
          text: category,
          data: "",
          items: subItems,
          active: subItems.some((item) => item.active),
        });
      }
    }

    for (const hiddenResource of hidden) {
      const gvrKey = gvrToKey(
        hiddenResource.group,
        hiddenResource.version,
        hiddenResource.resource,
      );
      const gvKey = gvToKey(hiddenResource.group, hiddenResource.version);

      apisGrouped.get(gvKey)?.delete(hiddenResource.resource);
      apisFlattened.delete(gvrKey);
    }

    const otherItems: SidebarItemConfig[] = [];

    for (const [gv, resources] of apisGrouped.entries()) {
      if (resources.size === 0) {
        continue;
      }

      const subItems: SidebarItemConfig[] = [];

      for (const [resource, { namespaced }] of resources.entries()) {
        const data = {
          group: keyToGv(gv).group,
          version: keyToGv(gv).version,
          resource: resource,
          namespaced: namespaced,
        };
        subItems.push({
          text: resource,
          data,
          items: [],
          active: isActive(data.group, data.version, data.resource),
        });
      }

      otherItems.push({
        text: gv,
        data: "",
        items: subItems,
        active: subItems.some((item) => item.active),
      });
    }

    if (otherItems.length > 0) {
      items.push({
        text: "Other",
        data: "",
        items: otherItems,
        active: otherItems.some((item) => item.active),
      });
    }

    if (items.every((item) => !item.active)) {
      group = "";
      version = "";
      resource = "";
    }
  }

  function onParamsChange() {
    items = [];

    (async () => {
      isLoadingSidebar = true;
      await sidebarRefresher?.refresh();
      isLoadingSidebar = false;
    })();
  }

  function updateActiveItem(item: SidebarItemConfig) {
    if (item.items.length > 0) {
      item.items.forEach(updateActiveItem);
      item.active = item.items.some((subItem) => subItem.active);
    } else if (item.data) {
      item.active = isActive(
        item.data.group,
        item.data.version,
        item.data.resource,
      );
    }
  }

  function onSelect(data: any) {
    ({ group, version, resource, namespaced } = data);
    items.forEach(updateActiveItem);
  }

  onMount(() => {
    sidebarRefresher = new Refresher({
      refresh: loadSidebar,
      onError: (e) => {
        ShowAlert("error", e.message);
      },
    });
    if (context) {
      onParamsChange();
    }
  });

  onDestroy(() => {
    sidebarRefresher?.abort();
  });
</script>

<ul
  class="sidebar list-unstyled mb-0 ps-3 py-3 overflow-y-auto border-end border-dark-subtle"
>
  {#if isLoadingSidebar}
    <li>Loading...</li>
  {:else if items.length === 0}
    <li>No resources found</li>
  {:else}
    {#each items.values() as item (item.text)}
      <SidebarItem config={item} select={onSelect} />
    {/each}
  {/if}
</ul>

<style lang="scss">
  .sidebar {
    min-width: 200px;
    max-width: 400px;
  }
</style>
