<script context="module" lang="ts">
  export type Item = { label: string; value: string };
</script>

<script lang="ts">
  import Fuse from "fuse.js";
  import { onMount } from "svelte";
  import { Dropdown as BSDropdown } from "bootstrap";

  export let isLoading = false;
  export let items: Item[] = [];
  export let selectedItem: string;
  export let noItemsMessage: string = "No items available";
  export let loadingMessage: string = "Loading items...";
  export let noSelectionMessage: string = "Select item...";
  export let alignEnd: boolean = false;

  let searchTerm: string = "";
  let dropDownItems: Item[] = [];
  let fuse = new Fuse(items, { keys: ["label"] });
  let selectedItemLabel = "";
  let navigationIndex = -1;
  let searchElement: HTMLInputElement;
  let dropdownElement: HTMLButtonElement;
  let bsDropdown: bootstrap.Dropdown | null;

  $: hasItems = items.length > 0;
  $: items, onItemsChanged();
  $: {
    if (isLoading) {
      bsDropdown?.hide();
    }
  }
  $: selectedItem && onSelectedItemChanged();

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
    disabled={!hasItems || isLoading}
    bind:this={dropdownElement}
    on:click={() => searchElement.focus()}
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
  <ul class="dropdown-menu pt-0" class:dropdown-menu-end={alignEnd}>
    <div class="mb-3 mt-0">
      <input
        type="text"
        class="form-control"
        placeholder="Search..."
        bind:this={searchElement}
        bind:value={searchTerm}
        on:input={onSearchUpdate}
        on:keydown={onSearchNavigate}
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
          on:mouseenter={() => (navigationIndex = i)}
          on:click={onItemSelect}>{dropDownItem.label}</a
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
