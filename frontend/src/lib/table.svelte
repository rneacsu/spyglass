<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import client from "./grpc/client";
  import type { ListResourceTabularReply } from "./grpc/proto/kube_pb";
  import { ShowAlert } from "./alerts.svelte";
  import { Refresher } from "./grpc/refresher";

  export let context: string;
  export let namespace: string;
  export let group: string;
  export let version: string;
  export let resource: string;
  export let namespaced: boolean;

  let table: ListResourceTabularReply | null = null;

  let isLoadingTable: boolean = false;

  let tableRefresher: Refresher | null = null;

  $: context, namespace, group, version, resource, onParamsChange();

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
        namespace: (!namespaced || namespace === "__all__") ? "" : namespace,
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
