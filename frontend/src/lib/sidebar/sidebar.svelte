<script lang="ts" module>
  export type GVRItemData = {
    group: string;
    version: string;
    resource: string;
    namespaced: boolean;
  };
</script>

<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import SidebarItem, { type SidebarItemConfig } from "./sidebarItem.svelte";
  import client from "../grpc/client";
  import { ShowAlert } from "../alerts.svelte";
  import { Refresher } from "../grpc/refresher";
  import { translate, translateResource } from "../translator";
  import { structure, hidden } from "./config";

  let {
    context = "",
    group = $bindable(""),
    version = $bindable(""),
    resource = $bindable(""),
    namespaced = $bindable(false),
  } = $props();

  let items: SidebarItemConfig<GVRItemData>[] = $state([]);
  let isLoadingSidebar: boolean = $state(false);

  let sidebarRefresher: Refresher | null = null;

  $effect(() => {
    context;
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

  function isActive(data: GVRItemData) {
    return (
      data.group === group &&
      data.version === version &&
      data.resource === resource
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
      const subItems: SidebarItemConfig<GVRItemData>[] = [];
      for (const categoryItem of categoryItems) {
        const gvKey = gvToKey(categoryItem.group, categoryItem.version);
        const gvrKey = gvrToKey(
          categoryItem.group,
          categoryItem.version,
          categoryItem.resource,
        );
        const apiItem = apisFlattened.get(gvrKey);

        if (apiItem) {
          const data: GVRItemData = {
            group: categoryItem.group,
            version: categoryItem.version,
            resource: categoryItem.resource,
            namespaced: apiItem.namespaced,
          };
          subItems.push({
            text: translateResource(gvrKey),
            items: [],
            data,
            active: isActive(data),
          });
          apisGrouped.get(gvKey)?.delete(categoryItem.resource);
          apisFlattened.delete(gvrKey);
        }
      }

      if (subItems.length > 0) {
        items.push({
          text: translate(category),
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

    const remainingItems: SidebarItemConfig<GVRItemData>[] = [];

    for (const [gv, resources] of apisGrouped.entries()) {
      if (resources.size === 0) {
        continue;
      }

      const subItems: SidebarItemConfig<GVRItemData>[] = [];

      for (const [resource, { namespaced }] of resources.entries()) {
        const data: GVRItemData = {
          group: keyToGv(gv).group,
          version: keyToGv(gv).version,
          resource: resource,
          namespaced: namespaced,
        };
        subItems.push({
          text: resource,
          data,
          items: [],
          active: isActive(data),
        });
      }

      remainingItems.push({
        text: gv,
        items: subItems,
        active: subItems.some((item) => item.active),
      });
    }

    if (items.length === 0) {
      items = remainingItems;
    } else {
      items[items.length - 1].items.push(...remainingItems);
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

  function updateActiveItem(item: SidebarItemConfig<GVRItemData>) {
    if (item.items.length > 0) {
      item.items.forEach(updateActiveItem);
      item.active = item.items.some((subItem) => subItem.active);
    } else if (item.data) {
      item.active = isActive(item.data);
    }
  }

  function onSelect(data?: GVRItemData) {
    if (!data) {
      return;
    }
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
    onParamsChange();
  });

  onDestroy(() => {
    sidebarRefresher?.abort();
  });
</script>

<ul class="menu overflow-y-auto border-e border-base-300 flex-nowrap w-72">
  {#if isLoadingSidebar}
    <li class="menu-disabled"><span>Loading...</span></li>
  {:else if items.length === 0}
    <li class="menu-disabled"><span>No resources found</span></li>
  {:else}
    {#each items.values() as item (item.text)}
      <SidebarItem config={item} select={onSelect} />
    {/each}
  {/if}
</ul>
