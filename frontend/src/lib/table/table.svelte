<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import client from "../grpc/client";
  import { ShowAlert } from "../alerts.svelte";
  import { Refresher } from "../grpc/refresher";
  import DataTable from "./dataTable.svelte";
  import type { ConfigColumns } from "datatables.net-bs5";
  import { getConfig } from "./config";
  import { translateTableColumn } from "$lib/translator";
  import { renderDefault, renderRelativeTime } from "./render";

  let {
    context = "",
    namespace = "",
    group = "",
    version = "",
    resource = "",
    namespaced = false,
  } = $props();

  let table: DataTable | null = null;
  let tableRefresher: Refresher | null = null;
  let recreateTable: boolean = false;
  let columnOrder: number[] = [];

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
      table?.init({});
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

    let tableConfig = getConfig(group, version, resource);

    let columns: ConfigColumns[];

    if (recreateTable) {
      columns = [];
      columnOrder = [];

      tableConfig.columnOrder.forEach((c) => {
        let index = data.columns.findIndex((col) => col.name === c);
        if (index !== -1) {
          columnOrder.push(index);
        }
      })
      data.columns.forEach((c, i) => {
        if (!columnOrder.includes(i)) {
          columnOrder.push(i);
        }
      });

      columns = columnOrder.map((i) => {
        const c = data.columns[i];
        return {
          title: c.name,
          visible: !tableConfig.hiddenColumns.includes(c.name),
          render: tableConfig.render[c.name],
        };
      });

      columns.unshift({ name: "Namespace", title: "Namespace", visible: namespaced && namespace === "__all__" });
      columns.unshift({ name: "Name", title: "Name", visible: tableConfig.showName });
      columns.unshift({ name: "Id", title: "Id", visible: false });
      columns.push({ name: "Age", title: "Age", visible: tableConfig.showAge, render: renderRelativeTime() });

      columns.forEach((c) => {
        // Translate columns
        c.title = translateTableColumn(c.title ?? "");
      });
    }

    const rowData = data.rows.map((r) => {

      // Reorder columns
      let row: any[] = columnOrder.map((i) => r.cells[i]);

      row.unshift(r.resource!.namespace);
      row.unshift(r.resource!.name);
      row.unshift(r.resource!.uid);
      row.push(r!.resource!.created!.seconds);

      return row;
    });

    if (recreateTable) {
      table?.init({
        columns: columns!,
        data: rowData,
        order: tableConfig.defaultOrder,
        columnDefs: [
          {
            targets: "_all",
            render: renderDefault(),
          },
        ],
      });
      recreateTable = false;
    } else {
      table?.replaceData(rowData);
    }
  }

  function onParamsChange() {
    recreateTable = true;

    (async () => {
      table?.processing(true);
      await tableRefresher?.refresh();
      table?.processing(false);
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
    background: var(--bs-body-bg);

    :global(table) {
      :global(th),
      :global(td) {
        :global(span) {
          max-width: 200px;
        }
      }
    }
  }
</style>
