<script lang="ts" module>
  type ResourceColumn = Column & {
    render?: RenderType;
  };

  function getRelativeTime(dateMs: number): string {
    const past = new Date(Number(dateMs));
    const diff = Date.now() - past.getTime();
    const future = diff < 0;

    const seconds = Math.floor(Math.abs(diff) / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    const years = Math.floor(days / 365);

    let result = "";

    if (years > 0) {
      result = `${years}y`;
    } else if (days > 0) {
      result = `${days}d`;
    } else if (hours > 0) {
      result = `${hours}h`;
    } else if (minutes > 0) {
      result = `${minutes}m`;
    } else {
      result = `${seconds}s`;
    }

    return future ? `in ${result}` : `${result} ago`;
  }
</script>

<script lang="ts">
  import { translateTableColumn } from "$lib/translator";
  import { onDestroy, onMount, untrack } from "svelte";
  import { ShowAlert } from "../alerts.svelte";
  import client from "../grpc/client";
  import { Refresher } from "../grpc/refresher";
  import { getConfig, type RenderType } from "./config";
  import DataTable, { type Column } from "./dataTable.svelte";

  let {
    context = "",
    namespace = "",
    group = "",
    version = "",
    resource = "",
    namespaced = false,
  } = $props();

  let tableRefresher: Refresher | null = null;
  let recreateTable: boolean = false;
  let isLoading = $state(false);
  let columns: ResourceColumn[] = $state([]);
  let rows: Record<string, any>[] = $state([]);
  let shouldDisplay = $derived(context && version && resource);
  let table: DataTable;

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

  function getStatusClass(status: string): string {
    const statusMap = {
      Completed: "badge-info",
      Succeeded: "badge-info",
      Running: "badge-success",
      Pending: "badge-warning",
      OOMKilled: "badge-error",
      Failed: "badge-error",
      CrashLoopBackOff: "badge-error",
    } as { [key: string]: string };

    return statusMap[status] || "badge-secondary";
  }

  async function loadTable(signal: AbortSignal) {
    if (!shouldDisplay) {
      rows = [];
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

    let columnOrder: number[] = [];

    tableConfig.columnOrder.forEach((c) => {
      let index = data.columns.findIndex((col) => col.name === c);
      if (index !== -1) {
        columnOrder.push(index);
      }
    });
    data.columns.forEach((c, i) => {
      if (!columnOrder.includes(i)) {
        columnOrder.push(i);
      }
    });

    if (recreateTable) {
      columns = columnOrder.map((i) => {
        const c = data.columns[i];
        return {
          name: c.name,
          sortable: true,
          hidden: tableConfig.hiddenColumns.includes(c.name),
          render: tableConfig.render[c.name],
        };
      });

      columns.unshift({
        name: "Namespace",
        sortable: true,
        hidden: !namespaced || namespace !== "__all__",
      });
      columns.unshift({
        name: "Name",
        sortable: true,
        hidden: !tableConfig.showName,
      });
      columns.unshift({
        name: "Id",
        sortable: true,
        hidden: true,
      });
      columns.push({
        name: "Age",
        sortable: true,
        hidden: !tableConfig.showAge,
        render: "timestamp",
      });

      columns.forEach((c) => {
        c.title = translateTableColumn(c.title ?? c.name);
      });

      table.setSort(tableConfig.defaultOrder);
      recreateTable = false;
    }

    rows = data.rows.map<Record<string, any>>((r) => {
      // Reorder columns
      let row: any[] = columnOrder.map((i) => r.cells[i]);

      row.unshift(r.resource!.namespace);
      row.unshift(r.resource!.name);
      row.unshift(r.resource!.uid);
      row.push(r!.resource!.created!.seconds);

      return Object.fromEntries(columns.map((col, i) => [col.name, row[i]]));
    });
  }

  function onParamsChange() {
    recreateTable = true;
    (async () => {
      isLoading = true;
      table.setSort();
      table.setFilter();
      await tableRefresher?.refresh();
      isLoading = false;
    })();
  }

  onMount(() => {
    tableRefresher = new Refresher({
      refresh: loadTable,
      onError: (e) => {
        ShowAlert("error", e.message);
      },
    });
    onParamsChange();
  });

  onDestroy(() => {
    tableRefresher?.abort();
  });
</script>

<DataTable bind:this={table} {isLoading} {columns} {rows} class="h-full">
  {#snippet cell(data, col)}
    {#if col.render === "timestamp"}
      <span title={new Date(Number(data) * 100).toISOString()}>
        {getRelativeTime(Number(data) * 1000)}
      </span>
    {:else if col.render === "selector"}
      {#each (data as string).split(",") as label}
        <span class="badge badge-soft badge-secondary me-1 mb-1">
          {label}
        </span>
      {/each}
    {:else if col.render === "status"}
      <span class={["badge", "badge-soft", getStatusClass(String(data))]}>
        {String(data)}
      </span>
    {:else}
      {String(data)}
    {/if}
  {/snippet}
</DataTable>
