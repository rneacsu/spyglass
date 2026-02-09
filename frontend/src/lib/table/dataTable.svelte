<script lang="ts" module>
  export type Column = {
    name: string;
    title?: string;
    sortable: boolean;
    hidden?: boolean;
  };
</script>

<script
  lang="ts"
  generics="TCell = any, TRow extends Record<string, TCell> = Record<string, TCell>, TColumn extends Column = Column"
>
  import {
    ChevronDown,
    ChevronUp,
    ChevronsUpDown,
    Search,
  } from "@lucide/svelte";

  import { type Snippet } from "svelte";

  let {
    isLoading = false,
    columns = [],
    rows = [],
    cell,
    sortBy = $bindable([]),
    filterBy = $bindable(""),
    ...props
  }: {
    isLoading?: boolean;
    rows?: TRow[];
    columns?: TColumn[];
    cell?: Snippet<[TCell, TColumn]>;
    sortBy?: { col: keyof TRow; dir: "asc" | "desc" }[];
    filterBy?: string;
    class?: string;
  } = $props();

  let numVisibleColumns: number = $derived(
    columns.filter((col) => !col.hidden).length,
  );

  let filteredRows: TRow[] = $derived.by(() => {
    return rows.filter((row) => {
      if (!filterBy) return true;
      return Object.values(row).some((cell) =>
        String(cell).toLowerCase().includes(filterBy.toLowerCase()),
      );
    });
  });

  let sortedRows: TRow[] = $derived.by(() => {
    let sorted = [...filteredRows];
    for (let { col, dir } of sortBy) {
      sorted.sort((a, b) => {
        if (a[col] < b[col]) return dir === "asc" ? -1 : 1;
        if (a[col] > b[col]) return dir === "asc" ? 1 : -1;
        return 0;
      });
    }
    return sorted;
  });

  let renderedRows: TRow[] = $derived(sortedRows);

  export function setSort(
    sort: { col: keyof TRow; dir: "asc" | "desc" }[] = [],
  ) {
    sortBy = sort;
  }

  export function setFilter(filter: string = "") {
    filterBy = filter;
  }
</script>

<div class={["flex flex-col", props.class]}>
  <label class="input my-2 self-end">
    <Search class="h-[1em] opacity-50" />
    <input bind:value={filterBy} type="search" required placeholder="Search" />
  </label>
  <div class="h-0 grow overflow-x-auto">
    {#if isLoading}
      <div class="flex items-center justify-center p-4">Loading data...</div>
    {:else}
      <table class="table bg-base-100 table-pin-rows">
        <thead>
          <tr>
            {#each columns as column}
              {#if !column.hidden}
                {#if column.sortable}
                  <th
                    class="cursor-pointer select-none group"
                    onclick={(event) => {
                      let existing = sortBy.find((s) => s.col === column.name);
                      if (!event.shiftKey) {
                        sortBy = existing ? [existing] : [];
                      }
                      if (existing) {
                        existing.dir = existing.dir === "asc" ? "desc" : "asc";
                      } else {
                        sortBy = [{ col: column.name, dir: "asc" }, ...sortBy];
                      }
                    }}
                  >
                    <div class="flex items-center">
                      <span class="me-1.5">{column.title ?? column.name}</span>
                      {#each sortBy.filter((s) => s.col === column.name) as sortByEntry}
                        {#if sortByEntry.dir === "asc"}
                          <ChevronUp class="inline h-[1em]" />
                        {:else}
                          <ChevronDown class="inline h-[1em]" />
                        {/if}
                      {:else}
                        <ChevronsUpDown
                          class="inline h-[1em] text-transparent group-hover:text-inherit"
                        />
                      {/each}
                    </div>
                  </th>
                {:else}
                  <th>{column.title ?? column.name}</th>
                {/if}
              {/if}
            {/each}
          </tr>
        </thead>
        <tbody>
          {#if renderedRows.length === 0}
            <tr>
              <td colspan={numVisibleColumns} class="text-center">
                No data available.
              </td>
            </tr>
          {:else}
            {#each renderedRows as row}
              <tr class="hover:bg-base-300">
                {#each columns as column}
                  {#if !column.hidden}
                    <td>
                      {#if cell}
                        {@render cell(row[column.name], column)}
                      {:else}
                        {String(row[column.name])}
                      {/if}
                    </td>
                  {/if}
                {/each}
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<style lang="scss">
</style>
