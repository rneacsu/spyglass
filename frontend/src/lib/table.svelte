<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import client from "./grpc/client";
  import type { ListResourceTabularReply } from "./grpc/proto/kube_pb";
  import { ShowAlert } from "./alerts.svelte";
  import { Refresher } from "./grpc/refresher";

  let {
    context = "",
    namespace = "",
    group = "",
    version = "",
    resource = "",
    namespaced = false,
  } = $props();

  let table: ListResourceTabularReply | null = $state(null);

  let isLoadingTable: boolean = $state(false);

  let tableRefresher: Refresher | null = null;

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
    if (!context || !version || !resource) {
      return;
    }

    ShowAlert("info", "Loading table...");

    table = await (
      await client
    ).listResourceTabular(
      {
        context,
        namespace: !namespaced || namespace === "__all__" ? "" : namespace,
        gvr: { group, version, resource },
      },
      { signal: signal },
    );
  }

  function onParamsChange() {
    table = null;

    (async () => {
      isLoadingTable = true;
      await tableRefresher?.refresh();
      isLoadingTable = false;
    })();
  }

  onMount(() => {
    tableRefresher = new Refresher({
      refresh: loadTable,
      onError: (e) => {
        ShowAlert("error", e.message);
      },
    });
  });

  onDestroy(() => {
    tableRefresher?.abort();
  });
</script>

<div class="h-100 overflow-y-auto">
  {#if isLoadingTable}
    <div class="mt-5 d-flex justify-content-center">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>
  {:else}
    <table class="table table-striped table-hover table-sm">
      <thead>
        <tr>
          {#if table}
            {#each table.columns as column}
              <th>{column.name}</th>
            {/each}
          {/if}
        </tr>
      </thead>
      <tbody>
        {#if table}
          {#if table.rows.length === 0}
            <tr>
              <td colspan={table.columns.length}>No items found</td>
            </tr>
          {:else}
            {#each table.rows as row}
              <tr>
                {#each row.cells as cell}
                  <td>{cell}</td>
                {/each}
              </tr>
            {/each}
          {/if}
        {/if}
      </tbody>
    </table>
  {/if}
</div>
