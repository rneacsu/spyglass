<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import client from "./grpc/client";
  import { ShowAlert } from "./alerts.svelte";
  import { Refresher } from "./grpc/refresher";
  import DataTable from "./datatable/dataTable.svelte";

  let {
    context = "",
    namespace = "",
    group = "",
    version = "",
    resource = "",
    namespaced = false,
  } = $props();

  let table: DataTable;
  let tableRefresher: Refresher | null = null;
  let recreateTable: boolean = false;

  let shouldDisplay = $derived(context && version && resource);

  $effect(() => {
    context !== null &&
      namespace !== null &&
      group !== null &&
      version !== null &&
      resource !== null &&
      untrack(() => {
        onParamsChange();
      });
  });

  async function loadTable(signal: AbortSignal) {
    if (!shouldDisplay) {
      return;
    }

    const data = await (
      await client
    ).listResourceTabular(
      {
        context,
        namespace: !namespaced || namespace === "__all__" ? "" : namespace,
        gvr: { group, version, resource },
      },
      { signal: signal },
    );

    const rowData = data.rows.map((r) => r.cells);

    if (recreateTable) {
      table.init({
        columns: data.columns.map((c) => ({ title: c.name })),
        data: rowData,
      });
      recreateTable = false;
    } else {
      table?.replaceData(rowData);
    }
  }

  function onParamsChange() {
    recreateTable = true;

    (async () => {
      table.processing(true);
      await tableRefresher?.refresh();
      table.processing(false);
    })();
  }

  onMount(() => {
    tableRefresher = new Refresher({
      refresh: loadTable,
      onError: (e) => {
        ShowAlert("error", e.message);
      },
    });
    if (shouldDisplay) {
      tableRefresher?.refresh();
    }
  });

  onDestroy(() => {
    tableRefresher?.abort();
  });
</script>

<div class="table-wrapper h-100">
  <DataTable bind:this={table} />
</div>

<style lang="scss">
  .table-wrapper {
    padding-bottom: 0.75rem;
  }
</style>
