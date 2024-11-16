<script lang="ts" module>
  export type Item = { label: string; value: string };
</script>

<script lang="ts">
  import Fuse from "fuse.js";
  import { onMount, untrack } from "svelte";
  import { Dropdown as BSDropdown } from "bootstrap";

  let {
    items = [],
    selectedItem = $bindable(),
    isLoading = false,
    noItemsMessage = "No items available",
    loadingMessage = "Loading items...",
    noSelectionMessage = "Select item...",
    alignEnd = false,
    disabled = false,
  } = $props();

  let searchTerm: string = $state("");
  let dropDownItems: Item[] = $state([]);
  let selectedItemLabel = $state("");
  let navigationIndex = $state(-1);

  let searchElement: HTMLInputElement;
  let dropdownElement: HTMLButtonElement;
  let bsDropdown: bootstrap.Dropdown | null;

  let fuse = new Fuse(items, { keys: ["label"] });

  let hasItems: boolean = $derived(items.length > 0);

  $effect(() => {
    items !== null &&
      untrack(() => {
        onItemsChanged();
      });
  });
  $effect(() => {
    isLoading &&
      untrack(() => {
        bsDropdown?.hide();
      });
  });
  $effect(() => {
    selectedItem &&
      untrack(() => {
        onSelectedItemChanged();
      });
  });

  function resetSearch() {
    searchTerm = "";
    onSearchUpdate();
  }

  function onSelectedItemChanged() {
    const item = findSelectedItem();
    selectedItemLabel = item?.label ?? "";
  }

  function findSelectedItem() {
    return items.find((item) => item.value === selectedItem);
  }

  function onItemsChanged() {
    fuse.setCollection(items);
    onSearchUpdate();

    if (!selectedItem && items.length > 0) {
      selectedItem = items[0].value;
    }

    if (selectedItem && !findSelectedItem()) {
      selectedItem = "";
    }
  }

  function onItemSelect(event: MouseEvent) {
    const target = event.target as HTMLAnchorElement;
    selectedItem = target.dataset.value ?? "";
  }

  function onSearchUpdate() {
    if (searchTerm === "") {
      dropDownItems = [...items];
      return;
    }
    dropDownItems = [...fuse.search(searchTerm).map((result) => result.item)];
    navigationIndex = -1;
  }

  function onSearchNavigate(event: KeyboardEvent) {
    if (event.key === "ArrowDown") {
      navigationIndex = Math.min(navigationIndex + 1, dropDownItems.length - 1);
    } else if (event.key === "ArrowUp") {
      navigationIndex = Math.max(navigationIndex - 1, 0);
    } else if (event.key === "Enter" && navigationIndex >= 0) {
      selectedItem = dropDownItems[navigationIndex].value;
      bsDropdown?.hide();
    }
  }

  onMount(() => {
    bsDropdown = new BSDropdown(dropdownElement);
    dropdownElement.addEventListener("hidden.bs.dropdown", () => {
      navigationIndex = -1;
      resetSearch();
    });
  });
</script>

<div class="dropdown">
  <button
    class="btn btn-secondary btn-sm dropdown-toggle text-truncate"
    data-bs-toggle="dropdown"
    type="button"
    aria-expanded="false"
    disabled={disabled || !hasItems || isLoading}
    bind:this={dropdownElement}
    onclick={() => searchElement.focus()}
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
  <ul class="dropdown-menu" class:dropdown-menu-end={alignEnd}>
    <div class="my-2 mx-3">
      <input
        type="text"
        class="form-control"
        placeholder="Search..."
        bind:this={searchElement}
        bind:value={searchTerm}
        oninput={onSearchUpdate}
        onkeydown={onSearchNavigate}
      />
    </div>
    {#if dropDownItems.length === 0}
      <li>
        <a class="dropdown-item disabled" href={"#"}>No items found</a>
      </li>
    {/if}
    {#each dropDownItems as dropDownItem, i}
      <li>
        <a
          class="dropdown-item text-truncate"
          class:active={navigationIndex === i}
          href={"#"}
          data-value={dropDownItem.value}
          data-index={i}
          onmouseenter={() => (navigationIndex = i)}
          onclick={onItemSelect}>{dropDownItem.label}</a
        >
      </li>
    {/each}
  </ul>
</div>

<style lang="scss">
  button.dropdown-toggle {
    max-width: 300px;
  }
  ul.dropdown-menu {
    max-height: 700px;
    overflow-y: auto;
    max-width: 550px;
    overflow-x: hidden;
  }
</style>
