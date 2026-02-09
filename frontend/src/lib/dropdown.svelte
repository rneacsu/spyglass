<script lang="ts" module>
  export type Item = { label: string; value: string };
</script>

<script lang="ts">
  import Fuse from "fuse.js";
  import { onMount, untrack } from "svelte";

  let {
    items = [],
    selectedItem = $bindable(),
    isLoading = false,
    noItemsMessage = "No items available",
    loadingMessage = "Loading items...",
    noSelectionMessage = "Select item...",
    ...props
  } = $props();

  let navigationIndex = $state(-1);
  let searchTerm: string = $state("");

  let fuse = new Fuse<any>([], { keys: ["label"] });

  let selectedItemLabel = $derived(
    items.find((item) => item.value === selectedItem)?.label ?? "",
  );
  let hasItems: boolean = $derived(items.length > 0);
  let dropDownItems: Item[] = $derived.by(() => {
    if (searchTerm === "") {
      return [...items];
    }
    return [...fuse.search(searchTerm).map((result) => result.item)];
  });
  let disabled: boolean = $derived(isLoading || !hasItems);

  let searchEl: HTMLInputElement;
  let dropdownEl: HTMLElement;

  $effect(() => {
    items !== null &&
      untrack(() => {
        onItemsChange();
      });
  });

  function onItemsChange() {
    fuse.setCollection(items);

    if (!selectedItem && items.length > 0) {
      selectedItem = items[0].value;
    }
  }

  function onSearchNavigate(event: KeyboardEvent) {
    if (event.key === "ArrowDown") {
      navigationIndex = Math.min(navigationIndex + 1, dropDownItems.length - 1);
    } else if (event.key === "ArrowUp") {
      navigationIndex = Math.max(navigationIndex - 1, 0);
    } else if (event.key === "Enter" && navigationIndex >= 0) {
      selectedItem = dropDownItems[navigationIndex].value;
      closeDropdown();
    }
  }

  function closeDropdown() {
    document.activeElement instanceof HTMLElement &&
      dropdownEl.contains(document.activeElement) &&
      document.activeElement.blur();
  }

  onMount(() => {
    onItemsChange();
  });
</script>

<div class={["dropdown", props.class]} bind:this={dropdownEl}>
  <button
    class={["btn m-1 max-w-52 block truncate", disabled && "btn-disabled"]}
    onfocus={() => {
      searchTerm = "";
      navigationIndex = -1;
      searchEl.focus();
    }}
  >
    {#if isLoading}
      {loadingMessage}
    {:else if !hasItems}
      {noItemsMessage}
    {:else if !selectedItem}
      {noSelectionMessage}
    {:else}
      {selectedItemLabel}
    {/if}
  </button>
  <ul
    class="menu dropdown-content bg-base-100 rounded-box z-5 w-52 p-2 shadow-sm"
  >
    <div class="my-2 mx-3">
      <input
        type="text"
        class="input"
        placeholder="Search..."
        bind:this={searchEl}
        bind:value={searchTerm}
        oninput={() => (navigationIndex = -1)}
        onkeydown={onSearchNavigate}
      />
    </div>
    {#if dropDownItems.length === 0}
      <li>
        <a href={"#"} class="menu-disabled">No items found</a>
      </li>
    {/if}
    {#each dropDownItems as dropDownItem, i}
      <li class="max-w-full">
        <a
          class={[
            "truncate block max-w-full",
            navigationIndex === i && "menu-focus",
            selectedItem === dropDownItem.value && "menu-active",
          ]}
          href={"#"}
          onmouseenter={() => (navigationIndex = i)}
          onclick={() => {
            selectedItem = dropDownItem.value;
            closeDropdown();
          }}>{dropDownItem.label}</a
        >
      </li>
    {/each}
  </ul>
</div>
