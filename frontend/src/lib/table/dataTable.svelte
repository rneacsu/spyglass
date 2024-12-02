<script lang="ts">
  import DataTable, { type Api, type Config } from "datatables.net-bs5";
  import "datatables.net-scroller-bs5";
  import "datatables.net-plugins/features/scrollResize/dataTables.scrollResize.mjs";
  import "datatables.net-plugins/dataRender/ellipsis";
  import { onDestroy, onMount } from "svelte";

  let table: Api<any>;
  let tableEl: HTMLTableElement;
  let theadEl: HTMLTableSectionElement;
  let tbodyEl: HTMLTableSectionElement;

  const defaultOptions: Config = {
    processing: true,
    columns: [],
    data: [],
    scrollY: "100px",
    scrollCollapse: true,
    scrollResize: true,
    scrollX: true,
    paging: false,
    // deferRender: true,
    // scroller: true,
  };

  export function init(config: Config) {
    table?.destroy();
    theadEl.innerHTML = "";
    tbodyEl.innerHTML = "";

    table = new DataTable(tableEl, { ...defaultOptions, ...config });
  }

  export function replaceData(data: any[]) {
    table.clear();
    table.rows.add(data);
    table.draw(false);
  }

  export function datatable() {
    return table;
  }

  export function processing(processing: boolean) {
    table.processing(processing);
  }

  onMount(() => {
    init({});
  });

  onDestroy(() => {
    table.destroy();
  });
</script>

<table class="table table-hover" bind:this={tableEl}>
  <thead bind:this={theadEl}> </thead>
  <tbody bind:this={tbodyEl}> </tbody>
</table>

<style lang="scss">
</style>
