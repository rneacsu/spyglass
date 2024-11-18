<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import client from "../grpc/client";
  import { ShowAlert } from "../alerts.svelte";
  import { Refresher } from "../grpc/refresher";
  import DataTable from "./dataTable.svelte";
  import type { ConfigColumns } from "datatables.net-bs5";
  import { overrides } from "./config";
  import dayjs from "dayjs";
  import relativeTime from "dayjs/plugin/relativeTime";
  import { translateTableColumn } from "$lib/translator";

  dayjs.extend(relativeTime);

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
  // let columnOrder: number[] = [];

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
      table.init({});
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

    let tableOverrides = overrides[`${group}/${version}::${resource}`];

    let columns: ConfigColumns[] | undefined;
    if (recreateTable) {
      // columnOrder = [];
      const specialColumns = ["Name", "Age", "Namespace"];
      columns = data.columns.map((c) => ({
        title: c.name,
        visible:
          !specialColumns.includes(c.name) &&
          !tableOverrides?.hiddenColumns.includes(c.name),
          render: tableOverrides?.render?.[c.name],
      }));

      if (namespaced) {
        columns!.unshift({ title: "Namespace" });
      }
      columns.unshift({ title: "Name" });
      columns!.push({ title: "Age" });

      columns.forEach((c) => {
        // Translate columns
        c.title = translateTableColumn(c.title ?? "");
      });
    }

    const rowData = data.rows.map((r) => {
      let row: string[] = r.cells;

      const age: string = dayjs
        .unix(Number(r!.resource!.created!.seconds))
        .fromNow();

      row.push(age);

      if (namespaced) {
        row.unshift(r.resource!.namespace);
      }
      row.unshift(r.resource!.name);

      return row;
    });

    if (recreateTable) {
      table.init({
        columns: columns,
        data: rowData,
      });
      recreateTable = false;
    } else {
      table.replaceData(rowData);
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
    background: var(--bs-body-bg);
  }
</style>
