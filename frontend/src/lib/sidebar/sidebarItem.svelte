<script lang="ts" module>
  export type SidebarItemConfig<T> = {
    text: string;
    data?: T;
    active: boolean;
    items: SidebarItemConfig<T>[];
  };
</script>

<script lang="ts" generics="T">
  import { onMount } from "svelte";

  import Self from "./sidebarItem.svelte";

  let {
    config = { text: "", active: false, items: [] },
    select = () => {},
  }: {
    config: SidebarItemConfig<T>;
    select: (data?: T) => void;
  } = $props();

  let isOpen: boolean = $state(false);

  onMount(() => {
    isOpen = config.active;
  });
</script>

<li class:items-start={config.items.length == 0} class="relative">
  {#if config.items.length > 0}
    <a
      class:menu-active={config.active}
      class="menu-dropdown-toggle"
      class:menu-dropdown-show={isOpen}
      onclick={() => (isOpen = config.active || !isOpen)}
      href={"#"}
    >
      <span class="truncate">{config.text}</span>
    </a>
    <ul class="menu-dropdown" class:menu-dropdown-show={isOpen}>
      {#each config.items as subItem (subItem.text)}
        <Self config={subItem} {select} />
      {/each}
    </ul>
  {:else}
    <a
      class:menu-active={config.active}
      onclick={() => select(config.data)}
      href={"#"}><span class="truncate">{config.text}</span></a
    >
  {/if}
</li>
